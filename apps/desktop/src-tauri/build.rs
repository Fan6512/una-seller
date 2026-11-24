fn main() {
    let manifest = tauri_build::AppManifest::new().commands(&[
        "desktop_app_version",
        "desktop_update_version",
        "official_tracking",
    ]);
    tauri_build::try_build(tauri_build::Attributes::new().app_manifest(manifest))
        .expect("failed to build desktop capabilities");
}
