import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string | Date, options?: Intl.DateTimeFormatOptions): string {
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  return dateObj.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    ...options,
  });
}

export function formatDateTime(date: string | Date): string {
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  return dateObj.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function getFileExtension(filename: string): string {
  return filename.slice((filename.lastIndexOf('.') - 1 >>> 0) + 2);
}

export function capitalizeFirst(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export function formatComponentType(type: string): string {
  return type.split('_').map(capitalizeFirst).join(' ');
}

export function getStatusColor(status: string): 'primary' | 'secondary' | 'default' | 'success' | 'warning' | 'danger' {
  const statusColors: Record<string, 'primary' | 'secondary' | 'default' | 'success' | 'warning' | 'danger'> = {
    operational: 'success',
    fault: 'danger',
    maintenance: 'warning',
    offline: 'default',
    pending: 'warning',
    processing: 'primary',
    completed: 'success',
    failed: 'danger',
    planned: 'default',
    in_progress: 'primary',
    cancelled: 'danger',
    on_hold: 'warning',
    requires_follow_up: 'secondary',
    low: 'default',
    medium: 'warning',
    high: 'secondary',
    critical: 'danger'
  };
  return statusColors[status] || 'default';
}

export function getProcessingStatusText(status: string): string {
  const statusText: Record<string, string> = {
    pending: 'Pending Processing',
    processing: 'Processing...',
    completed: 'Processed',
    failed: 'Processing Failed'
  };
  return statusText[status] || status;
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout;
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

export function formatConfidenceScore(score?: number): string {
  if (!score) return 'N/A';
  return `${(score * 100).toFixed(0)}%`;
}

export function getDocumentTypeColor(type: string): 'primary' | 'secondary' | 'default' | 'success' | 'warning' | 'danger' {
  const typeColors: Record<string, 'primary' | 'secondary' | 'default' | 'success' | 'warning' | 'danger'> = {
    field_service_report: 'primary',
    email: 'secondary',
    meeting_transcript: 'warning',
    work_order: 'success',
    inspection_report: 'primary',
    warranty_claim: 'danger',
    contract: 'secondary',
    manual: 'default',
    drawing: 'warning',
    other: 'default'
  };
  return typeColors[type] || 'default';
}

export function generateId(): string {
  return Math.random().toString(36).substring(2) + Date.now().toString(36);
}

export function parseErrorMessage(error: any): string {
  if (error?.response?.data?.message) {
    return error.response.data.message;
  }
  if (error?.message) {
    return error.message;
  }
  return 'An unexpected error occurred';
}

// Convert snake_case to camelCase
export function toCamelCase(obj: any): any {
  if (obj === null || obj === undefined) {
    return obj;
  }

  if (obj instanceof Date) {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(item => toCamelCase(item));
  }

  if (typeof obj === 'object') {
    const newObj: any = {};
    Object.keys(obj).forEach(key => {
      const camelKey = key.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
      newObj[camelKey] = toCamelCase(obj[key]);
    });
    return newObj;
  }

  return obj;
}