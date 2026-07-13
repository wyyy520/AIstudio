#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::process::{Command, Child};
use std::sync::Mutex;
use tauri::Manager;

struct BackendProcess(Mutex<Option<Child>>);

#[tauri::command]
fn get_backend_url() -> String {
    "http://localhost:8081".to_string()
}

fn start_backend() -> Option<Child> {
    let backend_path = std::env::current_dir()
        .ok()?
        .parent()?
        .parent()?
        .join("apps")
        .join("backend");

    let child = Command::new("go")
        .args(["run", "./cmd/"])
        .current_dir(&backend_path)
        .spawn()
        .ok()?;

    Some(child)
}

fn main() {
    let backend = start_backend();

    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(BackendProcess(Mutex::new(backend)))
        .setup(|app| {
            let handle = app.handle().clone();
            if let Some(window) = app.get_webview_window("main") {
                window.on_window_event(move |event| {
                    if let tauri::WindowEvent::CloseRequested { .. } = event {
                        if let Some(state) = handle.try_state::<BackendProcess>() {
                            if let Ok(mut child) = state.0.lock() {
                                if let Some(mut c) = child.take() {
                                    let _ = c.kill();
                                    let _ = c.wait();
                                }
                            }
                        }
                    }
                });
            }
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![get_backend_url])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
