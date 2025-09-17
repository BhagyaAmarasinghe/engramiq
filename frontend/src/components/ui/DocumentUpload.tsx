'use client';

import { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { 
  Card, 
  CardBody, 
  Button, 
  Progress, 
  Select, 
  SelectItem,
  Input,
  Textarea,
  Chip
} from '@heroui/react';
import { 
  IconUpload, 
  IconFile, 
  IconX, 
  IconCheck,
  IconAlertCircle
} from '@tabler/icons-react';
import { cn, formatFileSize } from '@/lib/utils';
import { DocumentType } from '@/types';

interface DocumentUploadProps {
  siteId: string;
  onUploadComplete?: (document: any) => void;
  onUploadError?: (error: string) => void;
  maxFiles?: number;
  maxSize?: number; // in bytes
  acceptedTypes?: string[];
  className?: string;
}

interface UploadFile extends File {
  id: string;
  progress: number;
  status: 'pending' | 'uploading' | 'completed' | 'error';
  error?: string;
  documentType?: DocumentType;
  metadata?: Record<string, any>;
}

const documentTypeOptions = [
  { key: 'field_service_report', label: 'Field Service Report' },
  { key: 'email', label: 'Email' },
  { key: 'meeting_transcript', label: 'Meeting Transcript' },
  { key: 'work_order', label: 'Work Order' },
  { key: 'inspection_report', label: 'Inspection Report' },
  { key: 'warranty_claim', label: 'Warranty Claim' },
  { key: 'contract', label: 'Contract' },
  { key: 'manual', label: 'Manual' },
  { key: 'drawing', label: 'Drawing' },
  { key: 'other', label: 'Other' },
];

export function DocumentUpload({
  siteId,
  onUploadComplete,
  onUploadError,
  maxFiles = 10,
  maxSize = 50 * 1024 * 1024, // 50MB
  acceptedTypes = ['.pdf', '.docx', '.doc', '.txt', '.eml'],
  className
}: DocumentUploadProps) {
  const [files, setFiles] = useState<UploadFile[]>([]);
  const [isUploading, setIsUploading] = useState(false);

  const onDrop = useCallback((acceptedFiles: File[], rejectedFiles: any[]) => {
    // Handle rejected files
    rejectedFiles.forEach(({ file, errors }) => {
      const errorMessage = errors.map((error: any) => {
        switch (error.code) {
          case 'file-too-large':
            return `File too large (max ${formatFileSize(maxSize)})`;
          case 'file-invalid-type':
            return 'File type not supported';
          default:
            return error.message;
        }
      }).join(', ');
      
      onUploadError?.(errorMessage);
    });

    // Add accepted files
    const newFiles: UploadFile[] = acceptedFiles.map((file) => ({
      ...file,
      id: Math.random().toString(36).substring(2),
      progress: 0,
      status: 'pending' as const,
      documentType: 'other' as DocumentType,
    }));

    setFiles(prev => [...prev, ...newFiles]);
  }, [maxSize, onUploadError]);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    maxFiles,
    maxSize,
    accept: acceptedTypes.reduce((acc, type) => {
      acc[`application/${type.replace('.', '')}`] = [type];
      return acc;
    }, {} as Record<string, string[]>),
    disabled: isUploading,
  });

  const removeFile = (fileId: string) => {
    setFiles(prev => prev.filter(f => f.id !== fileId));
  };

  const updateFileMetadata = (fileId: string, field: string, value: any) => {
    setFiles(prev => prev.map(f => 
      f.id === fileId 
        ? { 
            ...f, 
            [field]: value,
            metadata: { ...f.metadata, [field]: value }
          }
        : f
    ));
  };

  const uploadFiles = async () => {
    if (files.length === 0) return;
    
    setIsUploading(true);
    
    for (const file of files) {
      if (file.status !== 'pending') continue;
      
      try {
        // Update status to uploading
        setFiles(prev => prev.map(f => 
          f.id === file.id ? { ...f, status: 'uploading' as const } : f
        ));

        // Simulate upload progress (replace with actual API call)
        const formData = new FormData();
        formData.append('file', file);
        formData.append('document_type', file.documentType || 'other');
        
        if (file.metadata) {
          Object.entries(file.metadata).forEach(([key, value]) => {
            formData.append(key, JSON.stringify(value));
          });
        }

        // Progress simulation (replace with actual upload)
        for (let progress = 0; progress <= 100; progress += 10) {
          await new Promise(resolve => setTimeout(resolve, 100));
          setFiles(prev => prev.map(f => 
            f.id === file.id ? { ...f, progress } : f
          ));
        }

        // Mark as completed
        setFiles(prev => prev.map(f => 
          f.id === file.id ? { ...f, status: 'completed' as const, progress: 100 } : f
        ));

        onUploadComplete?.(file);
        
      } catch (error: any) {
        setFiles(prev => prev.map(f => 
          f.id === file.id 
            ? { ...f, status: 'error' as const, error: error.message }
            : f
        ));
        onUploadError?.(error.message);
      }
    }
    
    setIsUploading(false);
  };

  const getStatusIcon = (status: UploadFile['status']) => {
    switch (status) {
      case 'completed':
        return <IconCheck className="w-4 h-4 text-success" />;
      case 'error':
        return <IconAlertCircle className="w-4 h-4 text-danger" />;
      case 'uploading':
        return <div className="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />;
      default:
        return <IconFile className="w-4 h-4 text-default-500" />;
    }
  };

  const getStatusColor = (status: UploadFile['status']) => {
    switch (status) {
      case 'completed':
        return 'success';
      case 'error':
        return 'danger';
      case 'uploading':
        return 'primary';
      default:
        return 'default';
    }
  };

  return (
    <div className={cn('space-y-6', className)}>
      {/* Upload Dropzone */}
      <Card className="border-2 border-dashed border-default-200">
        <CardBody>
          <div
            {...getRootProps()}
            className={cn(
              'p-8 text-center cursor-pointer transition-colors',
              isDragActive && 'bg-primary/10 border-primary',
              isUploading && 'pointer-events-none opacity-50'
            )}
          >
            <input {...getInputProps()} />
            <div className="flex flex-col items-center gap-4">
              <div className="p-4 rounded-full bg-primary/10">
                <IconUpload className="w-8 h-8 text-primary" />
              </div>
              
              <div>
                <h3 className="text-lg font-semibold mb-2">
                  {isDragActive ? 'Drop files here' : 'Upload Documents'}
                </h3>
                <p className="text-default-500 mb-2">
                  Drag and drop files here, or click to select files
                </p>
                <p className="text-sm text-default-400">
                  Supported formats: {acceptedTypes.join(', ')} • Max {formatFileSize(maxSize)} per file
                </p>
              </div>
              
              <Button 
                color="primary" 
                variant="flat"
                disabled={isUploading}
              >
                Select Files
              </Button>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* File List */}
      {files.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-lg font-semibold">Files to Upload</h4>
          
          {files.map((file) => (
            <Card key={file.id} className="glass-effect">
              <CardBody className="p-4">
                <div className="flex items-start gap-4">
                  <div className="flex items-center gap-3 flex-1 min-w-0">
                    {getStatusIcon(file.status)}
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{file.name}</p>
                      <div className="flex items-center gap-2 text-sm text-default-500">
                        <span>{formatFileSize(file.size)}</span>
                        <Chip 
                          size="sm" 
                          color={getStatusColor(file.status)}
                          variant="flat"
                        >
                          {file.status}
                        </Chip>
                      </div>
                      {file.error && (
                        <p className="text-sm text-danger mt-1">{file.error}</p>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      isIconOnly
                      size="sm"
                      variant="light"
                      color="danger"
                      onClick={() => removeFile(file.id)}
                      disabled={file.status === 'uploading'}
                    >
                      <IconX className="w-4 h-4" />
                    </Button>
                  </div>
                </div>

                {file.status === 'uploading' && (
                  <Progress 
                    value={file.progress} 
                    color="primary" 
                    className="mt-3"
                    size="sm"
                  />
                )}

                {/* Document Type and Metadata */}
                {file.status === 'pending' && (
                  <div className="mt-4 space-y-3 pt-3 border-t border-default-200">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <Select
                        label="Document Type"
                        placeholder="Select document type"
                        selectedKeys={file.documentType ? [file.documentType] : []}
                        onSelectionChange={(keys) => {
                          const type = Array.from(keys)[0] as DocumentType;
                          updateFileMetadata(file.id, 'documentType', type);
                        }}
                        size="sm"
                      >
                        {documentTypeOptions.map((option) => (
                          <SelectItem key={option.key}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </Select>

                      <Input
                        label="Author"
                        placeholder="Document author"
                        size="sm"
                        value={file.metadata?.author || ''}
                        onValueChange={(value) => updateFileMetadata(file.id, 'author', value)}
                      />
                    </div>

                    <Textarea
                      label="Description"
                      placeholder="Optional description or notes"
                      size="sm"
                      maxRows={2}
                      value={file.metadata?.description || ''}
                      onValueChange={(value) => updateFileMetadata(file.id, 'description', value)}
                    />
                  </div>
                )}
              </CardBody>
            </Card>
          ))}

          {/* Upload Button */}
          <div className="flex justify-end">
            <Button
              color="primary"
              size="lg"
              onClick={uploadFiles}
              disabled={isUploading || files.every(f => f.status !== 'pending')}
              className="min-w-32"
            >
              {isUploading ? 'Uploading...' : 'Upload Files'}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}