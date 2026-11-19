#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

#[tauri::command]
fn desktop_update_version(app: tauri::AppHandle) -> Option<String> {
    (cfg!(all(windows, target_arch = "x86_64")) && app.config().plugins.0.contains_key("updater"))
        .then(|| app.package_info().version.to_string())
}

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_updater::Builder::new().build())
        .invoke_handler(tauri::generate_handler![desktop_update_version])
        .run(tauri::generate_context!())
        .expect("failed to start Una Seller");
}
