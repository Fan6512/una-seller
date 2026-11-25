#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod tracking;

// Installed package version is always available, even for unsigned QA builds.
#[tauri::command]
fn desktop_app_version(app: tauri::AppHandle) -> String {
    app.package_info().version.to_string()
}

#[tauri::command]
fn desktop_update_version(app: tauri::AppHandle) -> Option<String> {
    (cfg!(all(windows, target_arch = "x86_64")) && app.config().plugins.0.contains_key("updater"))
        .then(|| app.package_info().version.to_string())
}

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            if desktop_update_version(app.handle().clone()).is_some() {
                app.handle()
                    .plugin(tauri_plugin_updater::Builder::new().build())?;
            }
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            desktop_app_version,
            desktop_update_version,
            tracking::official_tracking
        ])
        .run(tauri::generate_context!())
        .expect("failed to start Una Seller");
}
