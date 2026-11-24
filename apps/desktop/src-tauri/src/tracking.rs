use tauri::Url;

fn tracking_url(value: &str) -> Result<Url, &'static str> {
    let invalid = "平台未提供有效的 HTTPS 物流查询地址";
    if value.len() > 2048 || value.contains('\\') || value.chars().any(char::is_control) {
        return Err(invalid);
    }
    let url = Url::parse(value).map_err(|_| invalid)?;
    let host = url.domain().ok_or(invalid)?;
    let tld = host.rsplit_once('.').map(|(_, tld)| tld).ok_or(invalid)?;
    if url.scheme() != "https"
        || url.port().is_some()
        || !url.username().is_empty()
        || url.password().is_some()
        || tld.len() < 2
        || !tld.bytes().all(|b| b.is_ascii_alphabetic())
        || [
            "localhost",
            "local",
            "internal",
            "test",
            "invalid",
            "example",
            "onion",
            "arpa",
        ]
        .contains(&tld)
    {
        return Err(invalid);
    }
    Ok(url)
}

fn same_carrier(url: &Url, origin: &Url) -> bool {
    tracking_url(url.as_str()).is_ok()
        && url
            .domain()
            .map(|host| host.strip_prefix("www.").unwrap_or(host))
            == origin
                .domain()
                .map(|host| host.strip_prefix("www.").unwrap_or(host))
}

#[tauri::command]
pub async fn official_tracking(
    app: tauri::AppHandle,
    url: String,
    carrier: String,
    external: bool,
) -> Result<(), String> {
    let url = tracking_url(&url)?;
    #[cfg(windows)]
    {
        if external {
            return windows::open_browser(&url).map_err(str::to_owned);
        }
        return windows::open_window(&app, url, &carrier);
    }
    #[cfg(not(windows))]
    {
        let _ = (app, url, carrier, external);
        Err("当前系统无法启动官方物流查询窗口".into())
    }
}

#[cfg(windows)]
mod windows {
    use super::{same_carrier, Url};
    use std::{
        io::Write,
        os::windows::process::CommandExt,
        process::{Command, Stdio},
        sync::{mpsc, Mutex},
        time::Duration,
    };
    use tauri::{
        menu::{MenuBuilder, SubmenuBuilder},
        webview::NewWindowResponse,
        AppHandle, Manager, WebviewUrl, WebviewWindowBuilder,
    };
    use webview2_com::{
        CoTaskMemPWSTR, Microsoft::Web::WebView2::Win32::COREWEBVIEW2_PERMISSION_STATE_DENY,
        NavigationCompletedEventHandler, PermissionRequestedEventHandler,
    };

    const LABEL: &str = "official-tracking";
    static OPEN_WINDOW_URL: Mutex<Option<Url>> = Mutex::new(None);

    #[link(name = "shell32")]
    extern "system" {
        fn ShellExecuteW(
            window: isize,
            operation: *const u16,
            file: *const u16,
            parameters: *const u16,
            directory: *const u16,
            show: i32,
        ) -> isize;
    }

    pub fn open_browser(url: &Url) -> Result<(), &'static str> {
        let verb: Vec<u16> = "open\0".encode_utf16().collect();
        let target: Vec<u16> = url.as_str().encode_utf16().chain([0]).collect();
        // Pass the validated URL directly to Windows. No command shell or string interpolation.
        let result = unsafe {
            ShellExecuteW(
                0,
                verb.as_ptr(),
                target.as_ptr(),
                std::ptr::null(),
                std::ptr::null(),
                1,
            )
        };
        (result > 32)
            .then_some(())
            .ok_or("无法打开浏览器，请复制地址后打开")
    }

    fn copy_address(url: &Url) -> std::io::Result<()> {
        let root = std::env::var_os("SystemRoot")
            .ok_or_else(|| std::io::Error::other("Windows system folder unavailable"))?;
        let mut child = Command::new(
            std::path::PathBuf::from(root)
                .join("System32")
                .join("clip.exe"),
        )
        .creation_flags(0x08000000) // CREATE_NO_WINDOW
        .stdin(Stdio::piped())
        .spawn()?;
        let mut input = child
            .stdin
            .take()
            .ok_or_else(|| std::io::Error::other("clipboard unavailable"))?;
        input.write_all(&[0xff, 0xfe])?;
        for unit in url.as_str().encode_utf16() {
            input.write_all(&unit.to_le_bytes())?;
        }
        drop(input);
        if child.wait()?.success() {
            Ok(())
        } else {
            Err(std::io::Error::other("clipboard unavailable"))
        }
    }

    pub fn open_window(app: &AppHandle, url: Url, carrier: &str) -> Result<(), String> {
        let mut opening = OPEN_WINDOW_URL
            .try_lock()
            .map_err(|_| "官方物流查询窗口正在打开")?;
        if let Some(window) = app.get_webview_window(LABEL) {
            let _ = window.set_focus();
            return if opening.as_ref() == Some(&url) {
                Ok(())
            } else {
                Err("请先关闭当前物流查询窗口，或在浏览器中打开".into())
            };
        }

        let carrier: String = carrier
            .chars()
            .filter(|c| !c.is_control())
            .take(60)
            .collect();
        let title = if carrier.trim().is_empty() {
            "官方物流查询".into()
        } else {
            format!("官方物流查询 · {carrier}")
        };
        let failed = format!("{title} · 加载失败，可在浏览器中打开");
        let menu_error = "无法创建官方物流查询菜单";
        let controls = SubmenuBuilder::new(app, "物流查询")
            .text("tracking-back", "后退")
            .text("tracking-forward", "前进")
            .text("tracking-reload", "刷新")
            .separator()
            .text("tracking-copy", "复制物流查询地址")
            .text("tracking-external", "在浏览器中打开")
            .text("tracking-close", "关闭")
            .build()
            .map_err(|_| menu_error)?;
        let menu = MenuBuilder::new(app)
            .item(&controls)
            .build()
            .map_err(|_| menu_error)?;

        let origin = url.clone();
        let menu_url = url.clone();
        let window = WebviewWindowBuilder::new(
            app,
            LABEL,
            WebviewUrl::External(Url::parse("about:blank").expect("static bootstrap URL")),
        )
        .title(&title)
        .inner_size(1000.0, 750.0)
        .min_inner_size(640.0, 480.0)
        .resizable(true)
        .visible(false)
        .incognito(true)
        .disable_drag_drop_handler()
        .on_download(|_, _| false)
        .on_navigation(move |next| same_carrier(next, &origin))
        .on_new_window(|_, _| NewWindowResponse::Deny)
        .menu(menu)
        .on_menu_event(move |host, event| {
            if host.label() != LABEL {
                return;
            }
            let Some(window) = host.app_handle().get_webview_window(LABEL) else {
                return;
            };
            let target = menu_url.clone();
            match event.id().as_ref() {
                "tracking-close" => {
                    let _ = window.close();
                }
                "tracking-copy" => {
                    std::thread::spawn(move || {
                        if copy_address(&target).is_err() {
                            let _ = window.set_title("官方物流查询 · 复制失败");
                        }
                    });
                }
                "tracking-external" => {
                    if let Err(error) = open_browser(&target) {
                        let _ = window.set_title(error);
                    }
                }
                action @ ("tracking-back" | "tracking-forward" | "tracking-reload") => {
                    let action = action.to_owned();
                    let _ = window.with_webview(move |view| unsafe {
                        if let Ok(core) = view.controller().CoreWebView2() {
                            let _ = match action.as_str() {
                                "tracking-back" => core.GoBack(),
                                "tracking-forward" => core.GoForward(),
                                _ => core.Reload(),
                            };
                        }
                    });
                }
                _ => {}
            }
        })
        .build()
        .map_err(|_| "无法创建官方物流查询窗口，可在浏览器中打开")?;

        let (ready, receiver) = mpsc::channel();
        let display = window.clone();
        let opened_url = url.clone();
        let initialize = window.with_webview(move |view| {
            // Apply restrictions before navigating away from the empty bootstrap document.
            let result = unsafe {
                (|| {
                    let core = view.controller().CoreWebView2()?;
                    let settings = core.Settings()?;
                    settings.SetIsWebMessageEnabled(false)?;
                    settings.SetAreHostObjectsAllowed(false)?;
                    settings.SetAreDevToolsEnabled(false)?;
                    let deny_permissions =
                        PermissionRequestedEventHandler::create(Box::new(|_, args| {
                            if let Some(args) = args {
                                args.SetState(COREWEBVIEW2_PERMISSION_STATE_DENY)?;
                            }
                            Ok(())
                        }));
                    let mut permission_registration = 0;
                    core.add_PermissionRequested(&deny_permissions, &mut permission_registration)?;
                    let completed =
                        NavigationCompletedEventHandler::create(Box::new(move |_, args| {
                            let mut succeeded = Default::default();
                            if let Some(args) = args {
                                args.IsSuccess(&mut succeeded)?;
                            }
                            let _ = display.set_title(if succeeded.as_bool() {
                                &title
                            } else {
                                &failed
                            });
                            Ok(())
                        }));
                    let mut registration = 0;
                    core.add_NavigationCompleted(&completed, &mut registration)?;
                    let target = CoTaskMemPWSTR::from(url.as_str());
                    core.Navigate(*target.as_ref().as_pcwstr())
                })()
            };
            let _ = ready.send(result.is_ok());
        });
        if initialize.is_err() || receiver.recv_timeout(Duration::from_secs(8)) != Ok(true) {
            let _ = window.close();
            return Err("官方物流查询窗口初始化失败，可在浏览器中打开".into());
        }
        if window.show().and_then(|_| window.set_focus()).is_err() {
            let _ = window.close();
            return Err("无法显示官方物流查询窗口，可在浏览器中打开".into());
        }
        *opening = Some(opened_url);
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn only_https_carrier_navigation() {
        let initial = tracking_url("https://www.carrier.com.br/track").unwrap();
        for url in [
            "https://carrier.com.br/result",
            "https://www.carrier.com.br/next",
        ] {
            assert!(same_carrier(&Url::parse(url).unwrap(), &initial));
        }
        for url in [
            "http://carrier.com.br",
            "https://carrier.com.br:8443",
            "https://carrier.com.br.evil.net",
            "https://alice@carrier.com.br",
            "https://127.0.0.1",
            "https://demo.localhost",
            "javascript:alert(1)",
            "data:text/html,Hello",
            "file:///C:/Windows",
            "about:blank",
        ] {
            assert!(!same_carrier(&Url::parse(url).unwrap(), &initial), "{url}");
        }
    }
}
