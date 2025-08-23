# Cloudlet Frontend

A modern, minimalist file management dashboard built with React, TypeScript, and Tailwind CSS.

## Features

- 📁 **File Management**: Browse, upload, download, and manage files and directories
- 🎯 **Drag & Drop**: Intuitive drag and drop file uploading
- 📊 **Smart Upload Strategy**: Automatically selects optimal upload method based on file size and count:
  - Single files → Single upload
  - Large files (>10MB) → Stream upload  
  - Many files or large batches → Batch upload
  - Multiple files → Multiple file upload
- 📋 **File Operations**: Create folders, rename, delete, and organize files
- 🎨 **Modern UI**: Built with shadcn/ui components for a clean, professional interface
- 📱 **Responsive Design**: Works seamlessly on desktop and mobile devices
- 🔄 **Real-time Updates**: Live feedback for all operations
- 📈 **Progress Tracking**: Visual progress indicators for file uploads

## Tech Stack

- **React 19** - Modern React with latest features
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first CSS framework
- **shadcn/ui** - High-quality, accessible UI components
- **Lucide React** - Beautiful icons
- **React Dropzone** - Drag and drop file uploads
- **Sonner** - Toast notifications
- **Vite** - Fast build tool and dev server

## Getting Started

### Prerequisites

- Node.js 18+ or Bun
- The Cloudlet backend server running on `http://localhost:8080`

### Installation

1. Install dependencies:
   ```bash
   bun install
   ```

2. Start the development server:
   ```bash
   bun run dev
   ```

3. Open your browser and navigate to `http://localhost:5173`

### Building for Production

```bash
bun run build
```

The built files will be output to `../web/` directory for serving by the Go backend.

## Usage

### File Management

- **Browse Files**: Click on folders to navigate, use breadcrumbs to go back
- **Upload Files**: Drag and drop files onto the upload zone or click to browse
- **Create Folders**: Click the "New Folder" button and enter a name
- **Rename Items**: Click the edit icon next to any file or folder
- **Delete Items**: Click the trash icon to delete files or folders
- **Download Files**: Click the download icon next to any file

### Upload Strategies

The application automatically selects the best upload method:

- **Single Upload**: For individual small files
- **Stream Upload**: For large files (>10MB) to handle them efficiently
- **Multiple Upload**: For several files at once
- **Batch Upload**: For many files (≥10) or large batches (≥500MB total)

### File Operations

All operations provide real-time feedback through toast notifications and immediate UI updates.

## API Integration

The frontend communicates with the Cloudlet backend using the following endpoints:

- `GET /api/v1/files[/{path}]` - List files and directories
- `POST /api/v1/upload` - Single file upload
- `POST /api/v1/upload/multiple` - Multiple file upload
- `POST /api/v1/upload/stream` - Stream upload for large files
- `POST /api/v1/upload/batch` - Batch upload with progress
- `POST /api/v1/directories` - Create directory
- `DELETE /api/v1/files` - Delete file or directory
- `POST /api/v1/rename` - Rename file or directory
- `GET /api/v1/download/{path}` - Download file

## Development

### Project Structure

```
src/
├── components/
│   ├── ui/                 # shadcn/ui components
│   ├── Dashboard.tsx       # Main dashboard component
│   ├── FileList.tsx       # File and directory listing
│   ├── FileUploadZone.tsx # Drag & drop upload area
│   ├── Breadcrumb.tsx     # Navigation breadcrumbs
│   └── *.tsx              # Other components
├── services/
│   ├── api.ts             # Base API configuration
│   ├── fileService.ts     # File operations API calls
│   ├── uploadService.ts   # Upload API calls  
│   └── uploadStrategy.ts  # Smart upload strategy pattern
├── lib/
│   ├── utils.ts           # Utility functions
│   └── formatters.ts      # Data formatting helpers
└── App.tsx                # Main app component
```

### Adding New Features

1. Create new components in `src/components/`
2. Add API calls to appropriate service files
3. Update TypeScript types in `src/services/api.ts`
4. Follow the existing patterns for error handling and user feedback

## Contributing

1. Follow the existing code style and patterns
2. Use TypeScript for all new code
3. Add proper error handling and user feedback
4. Test all functionality with the backend API
5. Update documentation for new features

## License

This project is part of Cloudlet and follows the same AGPL-3.0 license.