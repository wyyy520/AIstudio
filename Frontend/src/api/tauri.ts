import { invoke } from '@tauri-apps/api/tauri'

export interface AppInfo {
  name: string
  version: string
  platform: string
  arch: string
  data_dir: string
  config_dir: string
  cache_dir: string
}

export interface SystemInfo {
  os_name: string
  os_version: string
  kernel_version: string
  hostname: string
  cpu_brand: string
  cpu_cores: number
  total_memory_gb: number
  used_memory_gb: number
  total_swap_gb: number
  used_swap_gb: number
}

export interface FileInfo {
  name: string
  path: string
  size: number
  is_dir: boolean
  extension: string
}

export async function getAppInfo(): Promise<AppInfo> {
  return invoke<AppInfo>('get_app_info')
}

export async function getSystemInfo(): Promise<SystemInfo> {
  return invoke<SystemInfo>('get_system_info')
}

export async function openFile(path: string): Promise<string> {
  return invoke<string>('open_file', { path })
}

export async function saveFile(path: string, content: string): Promise<void> {
  return invoke<void>('save_file', { path, content })
}

export async function readDirectory(path: string): Promise<FileInfo[]> {
  return invoke<FileInfo[]>('read_directory', { path })
}

export async function createDirectory(path: string): Promise<void> {
  return invoke<void>('create_directory', { path })
}

export async function removeFile(path: string): Promise<void> {
  return invoke<void>('remove_file', { path })
}

export async function renameItem(oldPath: string, newPath: string): Promise<void> {
  return invoke<void>('rename_item', { oldPath, newPath })
}

export async function pathExists(path: string): Promise<boolean> {
  return invoke<boolean>('path_exists', { path })
}

export async function selectDirectory(): Promise<string | null> {
  return invoke<string | null>('select_directory')
}

export async function openFileDialog(): Promise<string | null> {
  return invoke<string | null>('open_file_dialog')
}

export async function saveFileDialog(defaultName: string): Promise<string | null> {
  return invoke<string | null>('save_file_dialog', { defaultName })
}

export async function getPlatform(): Promise<string> {
  const info = await getAppInfo()
  return info.platform
}

export async function isTauri(): Promise<boolean> {
  try {
    await getAppInfo()
    return true
  } catch {
    return false
  }
}
