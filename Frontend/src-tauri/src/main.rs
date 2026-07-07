#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::Serialize;
use std::fs;
use std::path::PathBuf;
use sysinfo::System;
use tauri::api::dialog::blocking::FileDialogBuilder;

#[derive(Serialize)]
struct AppInfo {
    name: String,
    version: String,
    platform: String,
    arch: String,
    data_dir: String,
    config_dir: String,
    cache_dir: String,
}

#[derive(Serialize)]
struct SystemInfo {
    os_name: String,
    os_version: String,
    kernel_version: String,
    hostname: String,
    cpu_brand: String,
    cpu_cores: usize,
    total_memory_gb: f64,
    used_memory_gb: f64,
    total_swap_gb: f64,
    used_swap_gb: f64,
}

#[derive(Serialize)]
struct FileInfo {
    name: String,
    path: String,
    size: u64,
    is_dir: bool,
    extension: String,
}

#[tauri::command]
fn get_app_info(app_handle: tauri::AppHandle) -> AppInfo {
    let package_info = app_handle.package_info();
    let resolver = app_handle.path_resolver();

    AppInfo {
        name: package_info.package_name().to_string(),
        version: package_info.version.to_string(),
        platform: std::env::consts::OS.to_string(),
        arch: std::env::consts::ARCH.to_string(),
        data_dir: resolver
            .app_data_dir()
            .unwrap_or_default()
            .to_string_lossy()
            .to_string(),
        config_dir: resolver
            .app_config_dir()
            .unwrap_or_default()
            .to_string_lossy()
            .to_string(),
        cache_dir: resolver
            .app_cache_dir()
            .unwrap_or_default()
            .to_string_lossy()
            .to_string(),
    }
}

#[tauri::command]
fn get_system_info() -> SystemInfo {
    let mut sys = System::new();
    sys.refresh_all();

    let cpu_brand = if sys.cpus().is_empty() {
        String::from("Unknown")
    } else {
        sys.cpus()[0].brand().to_string()
    };

    SystemInfo {
        os_name: System::name().unwrap_or_default(),
        os_version: System::os_version().unwrap_or_default(),
        kernel_version: System::kernel_version().unwrap_or_default(),
        hostname: System::host_name().unwrap_or_default(),
        cpu_brand,
        cpu_cores: sys.cpus().len(),
        total_memory_gb: sys.total_memory() as f64 / 1024.0 / 1024.0 / 1024.0,
        used_memory_gb: sys.used_memory() as f64 / 1024.0 / 1024.0 / 1024.0,
        total_swap_gb: sys.total_swap() as f64 / 1024.0 / 1024.0 / 1024.0,
        used_swap_gb: sys.used_swap() as f64 / 1024.0 / 1024.0 / 1024.0,
    }
}

#[tauri::command]
fn open_file(path: String) -> Result<String, String> {
    fs::read_to_string(&path).map_err(|e| format!("Failed to read file: {}", e))
}

#[tauri::command]
fn save_file(path: String, content: String) -> Result<(), String> {
    if let Some(parent) = PathBuf::from(&path).parent() {
        fs::create_dir_all(parent).map_err(|e| format!("Failed to create directory: {}", e))?;
    }
    fs::write(&path, &content).map_err(|e| format!("Failed to write file: {}", e))
}

#[tauri::command]
fn read_directory(path: String) -> Result<Vec<FileInfo>, String> {
    let entries = fs::read_dir(&path).map_err(|e| format!("Failed to read directory: {}", e))?;

    let mut files = Vec::new();
    for entry in entries {
        let entry = entry.map_err(|e| format!("Failed to read entry: {}", e))?;
        let metadata = entry
            .metadata()
            .map_err(|e| format!("Failed to read metadata: {}", e))?;
        let path = entry.path();

        files.push(FileInfo {
            name: entry.file_name().to_string_lossy().to_string(),
            path: path.to_string_lossy().to_string(),
            size: metadata.len(),
            is_dir: metadata.is_dir(),
            extension: path
                .extension()
                .map(|e| e.to_string_lossy().to_string())
                .unwrap_or_default(),
        });
    }

    files.sort_by(|a, b| a.name.to_lowercase().cmp(&b.name.to_lowercase()));
    Ok(files)
}

#[tauri::command]
fn create_directory(path: String) -> Result<(), String> {
    fs::create_dir_all(&path).map_err(|e| format!("Failed to create directory: {}", e))
}

#[tauri::command]
fn remove_file(path: String) -> Result<(), String> {
    let p = PathBuf::from(&path);
    if p.is_dir() {
        fs::remove_dir_all(&p).map_err(|e| format!("Failed to remove directory: {}", e))
    } else {
        fs::remove_file(&p).map_err(|e| format!("Failed to remove file: {}", e))
    }
}

#[tauri::command]
fn rename_item(old_path: String, new_path: String) -> Result<(), String> {
    fs::rename(&old_path, &new_path).map_err(|e| format!("Failed to rename: {}", e))
}

#[tauri::command]
fn path_exists(path: String) -> bool {
    PathBuf::from(&path).exists()
}

#[tauri::command]
fn select_directory() -> Result<Option<String>, String> {
    let path = FileDialogBuilder::new()
        .pick_folder()
        .map(|p| p.to_string_lossy().to_string());
    Ok(path)
}

#[tauri::command]
fn open_file_dialog() -> Result<Option<String>, String> {
    let path = FileDialogBuilder::new()
        .pick_file()
        .map(|p| p.to_string_lossy().to_string());
    Ok(path)
}

#[tauri::command]
fn save_file_dialog(default_name: String) -> Result<Option<String>, String> {
    let path = FileDialogBuilder::new()
        .set_file_name(&default_name)
        .save_file()
        .map(|p| p.to_string_lossy().to_string());
    Ok(path)
}

fn main() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            get_app_info,
            get_system_info,
            open_file,
            save_file,
            read_directory,
            create_directory,
            remove_file,
            rename_item,
            path_exists,
            select_directory,
            open_file_dialog,
            save_file_dialog,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
