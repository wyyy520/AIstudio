#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::Serialize;
use std::ffi::CString;
use std::fs;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::path::PathBuf;
use std::process::{Child, Command, Stdio};
use std::sync::Mutex;
use std::time::{Duration, Instant};
use sysinfo::System;
use tauri::api::dialog::blocking::FileDialogBuilder;
use tauri::Manager;

// ============================================================================
// Windows FFI
// ============================================================================

#[cfg(windows)]
extern "system" {
    fn MessageBoxA(hWnd: *mut std::ffi::c_void, lpText: *const std::ffi::c_char, lpCaption: *const std::ffi::c_char, uType: u32) -> i32;
    fn GetLocalTime(lpSystemTime: *mut SYSTEMTIME);
}

#[cfg(windows)]
#[repr(C)]
#[allow(non_snake_case)]
struct SYSTEMTIME {
    wYear: u16,
    wMonth: u16,
    wDayOfWeek: u16,
    wDay: u16,
    wHour: u16,
    wMinute: u16,
    wSecond: u16,
    wMilliseconds: u16,
}

const MB_OK: u32 = 0x00000000;
const MB_ICONERROR: u32 = 0x00000010;

// ============================================================================
// Error Dialog
// ============================================================================

fn show_error_box(title: &str, message: &str) {
    #[cfg(windows)]
    unsafe {
        let title_c = CString::new(title).unwrap_or_else(|_| CString::new("Error").unwrap());
        let msg_c = CString::new(message).unwrap_or_else(|_| CString::new("Unknown error").unwrap());
        MessageBoxA(std::ptr::null_mut(), msg_c.as_ptr(), title_c.as_ptr(), MB_OK | MB_ICONERROR);
    }
    #[cfg(not(windows))]
    {
        eprintln!("[ERROR] {}: {}", title, message);
    }
}

// ============================================================================
// Timestamp (Windows local time)
// ============================================================================

fn format_timestamp() -> String {
    #[cfg(windows)]
    unsafe {
        let mut st = std::mem::zeroed::<SYSTEMTIME>();
        GetLocalTime(&mut st);
        return format!("{:04}-{:02}-{:02} {:02}:{:02}:{:02}.{:03}",
            st.wYear, st.wMonth, st.wDay, st.wHour, st.wMinute, st.wSecond, st.wMilliseconds);
    }
    #[cfg(not(windows))]
    {
        use std::time::{SystemTime, UNIX_EPOCH};
        let dur = SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default();
        format!("{}", dur.as_secs())
    }
}

// ============================================================================
// Directory Helpers
// ============================================================================

fn get_app_dir() -> PathBuf {
    std::env::current_exe()
        .ok()
        .and_then(|p| p.parent().map(|p| p.to_path_buf()))
        .unwrap_or_else(|| PathBuf::from("."))
}

fn find_root_dir() -> PathBuf {
    let exe_dir = get_app_dir();
    let mut dir = exe_dir.clone();
    for _ in 0..6 {
        if dir.join("Config").join("app.yaml").is_file() {
            return dir;
        }
        if let Some(parent) = dir.parent() {
            dir = parent.to_path_buf();
        } else {
            break;
        }
    }
    log_to_console(&format!("[startup] WARNING: Could not find root dir from {}", exe_dir.display()));
    exe_dir
}

fn log_to_console(msg: &str) {
    // In debug mode this shows in the console; in release it's suppressed
    eprintln!("{}", msg);
}

// ============================================================================
// Startup Log
// ============================================================================

struct StartupLog {
    log_dir: PathBuf,
    start: Instant,
}

impl StartupLog {
    fn new(root_dir: &PathBuf) -> Self {
        let log_dir = root_dir.join("Runtime").join("logs");
        let _ = fs::create_dir_all(&log_dir);
        let sl = StartupLog {
            log_dir,
            start: Instant::now(),
        };
        sl.write("=== AIStudio Startup Log ===");
        sl.write(&format!("Timestamp: {}", format_timestamp()));
        sl
    }

    fn write(&self, msg: &str) {
        let elapsed = self.start.elapsed();
        let prefix = format!("[{:07.3}s]", elapsed.as_secs_f64());
        let entry = format!("{} {}", prefix, msg);
        let path = self.log_dir.join("startup.log");
        if let Ok(mut f) = fs::OpenOptions::new().create(true).append(true).open(path) {
            let _ = writeln!(f, "{}", entry);
        }
        log_to_console(&format!("[startup] {}", entry));
    }

    fn write_error(&self, msg: &str) {
        self.write(&format!("[ERROR] {}", msg));
    }

    fn path(&self) -> PathBuf {
        self.log_dir.join("startup.log")
    }
}

// ============================================================================
// Health Check
// ============================================================================

fn check_health(host: &str, port: u16, path: &str) -> bool {
    let addr = format!("{}:{}", host, port);
    let mut stream = match TcpStream::connect_timeout(
        &addr.parse().unwrap(),
        Duration::from_secs(3),
    ) {
        Ok(s) => s,
        Err(_) => return false,
    };
    let _ = stream.set_read_timeout(Some(Duration::from_secs(3)));

    let request = format!(
        "GET {} HTTP/1.0\r\nHost: {}:{}\r\nConnection: close\r\n\r\n",
        path, host, port
    );
    if stream.write_all(request.as_bytes()).is_err() {
        return false;
    }
    let mut response = String::new();
    if stream.read_to_string(&mut response).is_err() {
        return false;
    }
    response.contains("200 OK")
}

// ============================================================================
// Process I/O Helpers
// ============================================================================

fn stdio_from_file(file: &Option<fs::File>) -> Stdio {
    file.as_ref()
        .and_then(|f| f.try_clone().ok())
        .map(Stdio::from)
        .unwrap_or(Stdio::null())
}

// ============================================================================
// Existing Tauri Commands
// ============================================================================

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
        data_dir: resolver.app_data_dir().unwrap_or_default().to_string_lossy().to_string(),
        config_dir: resolver.app_config_dir().unwrap_or_default().to_string_lossy().to_string(),
        cache_dir: resolver.app_cache_dir().unwrap_or_default().to_string_lossy().to_string(),
    }
}

#[tauri::command]
fn get_system_info() -> SystemInfo {
    let mut sys = System::new();
    sys.refresh_all();

    SystemInfo {
        os_name: System::name().unwrap_or_default(),
        os_version: System::os_version().unwrap_or_default(),
        kernel_version: System::kernel_version().unwrap_or_default(),
        hostname: System::host_name().unwrap_or_default(),
        cpu_brand: if sys.cpus().is_empty() { "Unknown".into() } else { sys.cpus()[0].brand().to_string() },
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
        let metadata = entry.metadata().map_err(|e| format!("Failed to read metadata: {}", e))?;
        let path = entry.path();
        files.push(FileInfo {
            name: entry.file_name().to_string_lossy().to_string(),
            path: path.to_string_lossy().to_string(),
            size: metadata.len(),
            is_dir: metadata.is_dir(),
            extension: path.extension().map(|e| e.to_string_lossy().to_string()).unwrap_or_default(),
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
    Ok(FileDialogBuilder::new().pick_folder().map(|p| p.to_string_lossy().to_string()))
}

#[tauri::command]
fn open_file_dialog() -> Result<Option<String>, String> {
    Ok(FileDialogBuilder::new().pick_file().map(|p| p.to_string_lossy().to_string()))
}

#[tauri::command]
fn save_file_dialog(default_name: String) -> Result<Option<String>, String> {
    Ok(FileDialogBuilder::new().set_file_name(&default_name).save_file().map(|p| p.to_string_lossy().to_string()))
}

// ============================================================================
// Tauri App State (holds child processes for cleanup)
// ============================================================================

struct AppChildren {
    backend: Mutex<Option<Child>>,
    engine: Mutex<Option<Child>>,
}

// ============================================================================
// Main
// ============================================================================

fn main() {
    let root_dir = find_root_dir();
    let log = StartupLog::new(&root_dir);
    let app_dir = get_app_dir();

    log.write(&format!("App executable: {}", app_dir.display()));
    log.write(&format!("Root directory: {}", root_dir.display()));
    log.write(&format!("OS: {} {}", std::env::consts::OS, std::env::consts::ARCH));
    log.write(&format!("Args: {:?}", std::env::args().collect::<Vec<_>>()));

    // ========================================================================
    // Phase 1: Start Backend
    // ========================================================================
    let backend_path = root_dir.join("Backend").join("cmd.exe");
    let backend_config = root_dir.join("Config").join("backend.yaml");
    let backend_dir = root_dir.join("Backend");

    if !backend_path.exists() {
        let msg = format!("Backend not found: {}", backend_path.display());
        log.write_error(&msg);
        show_error_box("AIStudio 启动失败",
            &format!("Backend 可执行文件不存在：{}\n\n请检查安装完整性。\n\n日志: {}", backend_path.display(), log.path().display()));
        return;
    }
    if !backend_config.exists() {
        let msg = format!("Backend config not found: {}", backend_config.display());
        log.write_error(&msg);
        show_error_box("AIStudio 启动失败",
            &format!("配置文件不存在：{}\n\n请检查安装完整性。\n\n日志: {}", backend_config.display(), log.path().display()));
        return;
    }

    // Create backend log file
    let backend_log_path = log.log_dir.join("backend.log");
    let backend_log = match fs::File::create(&backend_log_path) {
        Ok(f) => {
            log.write(&format!("Backend log: {}", backend_log_path.display()));
            Some(f)
        }
        Err(e) => {
            log.write(&format!("Warning: could not create backend log: {}", e));
            None
        }
    };

    log.write(&format!("Starting Backend: {}", backend_path.display()));
    log.write(&format!("Backend config: {}", backend_config.display()));
    log.write(&format!("Backend CWD: {}", backend_dir.display()));

    let mut backend_child = match Command::new(&backend_path)
        .env("AISTUDIO_CONFIG", backend_config.to_str().unwrap())
        .current_dir(&backend_dir)
        .stdout(stdio_from_file(&backend_log))
        .stderr(stdio_from_file(&backend_log))
        .stdin(Stdio::null())
        .spawn()
    {
        Ok(child) => {
            log.write(&format!("Backend started, PID: {}", child.id()));
            log.write(&format!("Backend log: {}", backend_log_path.display()));
            child
        }
        Err(e) => {
            let msg = format!("Backend start failed: {}", e);
            log.write_error(&msg);
            show_error_box("AIStudio 启动失败",
                &format!("Backend 启动失败：{}\n\n日志: {}", e, log.path().display()));
            return;
        }
    };

    // ========================================================================
    // Phase 2: Backend Health Check
    // ========================================================================
    log.write("Waiting for Backend health check...");
    let deadline = Instant::now() + Duration::from_secs(30);
    let mut healthy = false;

    while Instant::now() < deadline {
        if check_health("127.0.0.1", 8081, "/api/health") {
            healthy = true;
            break;
        }
        std::thread::sleep(Duration::from_millis(500));
    }

    if !healthy {
        let _ = backend_child.kill();
        let _ = backend_child.wait();
        let msg = "Backend health check timeout (30s)".to_string();
        log.write_error(&msg);
        show_error_box("AIStudio 启动失败",
            &format!("Backend 启动超时（30秒），请检查：\n1. 端口 8081 是否被占用\n2. 防火墙是否阻止\n3. 系统资源是否充足\n\n日志: {}", log.path().display()));
        return;
    }

    log.write("Backend health check passed (200 OK)");

    // ========================================================================
    // Phase 3: Start Python Engine (optional)
    // ========================================================================
    let engine_path = root_dir.join("Engine").join("server.py");
    let engine_child: Option<Child> = if engine_path.exists() {
        let engine_log_path = log.log_dir.join("engine.log");
        let engine_log = match fs::File::create(&engine_log_path) {
            Ok(f) => {
                log.write(&format!("Engine log: {}", engine_log_path.display()));
                Some(f)
            }
            Err(e) => {
                log.write(&format!("Warning: could not create engine log: {}", e));
                None
            }
        };
        log.write(&format!("Starting Engine: {}", engine_path.display()));

        match Command::new("python")
            .arg(engine_path.to_str().unwrap())
            .arg("--port")
            .arg("8082")
            .arg("--host")
            .arg("127.0.0.1")
            .current_dir(root_dir.join("Engine"))
            .env("PYTHONUNBUFFERED", "1")
            .stdout(stdio_from_file(&engine_log))
            .stderr(stdio_from_file(&engine_log))
            .stdin(Stdio::null())
            .spawn()
        {
            Ok(child) => {
                log.write(&format!("Engine started, PID: {}", child.id()));
                log.write(&format!("Engine log: {}", engine_log_path.display()));
                Some(child)
            }
            Err(e) => {
                log.write(&format!("Engine start skipped (non-fatal): {}", e));
                None
            }
        }
    } else {
        log.write(&format!("Engine script not found (skipped): {}", engine_path.display()));
        None
    };

    // ========================================================================
    // Phase 4: Start Tauri GUI
    // ========================================================================
    log.write("Starting Tauri GUI...");
    log.write("============================================================");

    let children = AppChildren {
        backend: Mutex::new(Some(backend_child)),
        engine: Mutex::new(engine_child),
    };

    let result = tauri::Builder::default()
        .manage(children)
        .setup(|app| {
            // Register window close handler to clean up child processes
            if let Some(window) = app.get_window("main") {
                let handle = app.handle();
                window.on_window_event(move |event| {
                    if let tauri::WindowEvent::CloseRequested { .. } = event {
                        cleanup_children(&handle);
                    }
                });
            }
            Ok(())
        })
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
        .build(tauri::generate_context!());

    match result {
        Ok(app) => {
            app.run(|_handle, _event| {});
        }
        Err(e) => {
            let msg = format!("Tauri initialization failed: {}", e);
            log.write_error(&msg);
            show_error_box("AIStudio 启动失败",
                &format!("GUI 初始化失败：{}\n\n请检查系统环境（需要 WebView2 运行时）。\n\n日志: {}", e, log.path().display()));
        }
    }

    // ========================================================================
    // Phase 5: Cleanup (kill child processes)
    // ========================================================================
    log.write("Shutting down...");
    // Children will be cleaned up when AppChildren is dropped
}

fn cleanup_children(handle: &tauri::AppHandle) {
    if let Some(state) = handle.try_state::<AppChildren>() {
        // Kill backend
        if let Ok(mut child) = state.backend.lock() {
            if let Some(mut c) = child.take() {
                let _ = c.kill();
                let _ = c.wait();
            }
        }
        // Kill engine
        if let Ok(mut child) = state.engine.lock() {
            if let Some(mut c) = child.take() {
                let _ = c.kill();
                let _ = c.wait();
            }
        }
    }
}
