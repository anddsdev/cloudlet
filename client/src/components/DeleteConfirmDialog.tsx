import React from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { AlertTriangle, Folder, File } from 'lucide-react';

interface DeleteConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  itemName: string;
  itemPath: string;
  isDirectory: boolean;
  onConfirm: (recursive?: boolean) => Promise<void>;
  isDeleting?: boolean;
  hasChildren?: boolean;
}

export const DeleteConfirmDialog: React.FC<DeleteConfirmDialogProps> = ({
  open,
  onOpenChange,
  itemName,
  itemPath,
  isDirectory,
  onConfirm,
  isDeleting = false,
  hasChildren = false,
}) => {
  const handleDelete = (recursive: boolean = false) => {
    onConfirm(recursive);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-destructive" />
            Confirm Deletion
          </DialogTitle>
          <DialogDescription>
            This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        
        <div className="py-4">
          <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg mb-4">
            {isDirectory ? (
              <Folder className="h-6 w-6 text-blue-500" />
            ) : (
              <File className="h-6 w-6 text-gray-500" />
            )}
            <div>
              <p className="font-medium">{itemName}</p>
              <p className="text-sm text-muted-foreground">{itemPath}</p>
            </div>
          </div>

          {isDirectory && hasChildren ? (
            <div className="space-y-3">
              <div className="flex items-start gap-2 p-3 border border-orange-200 bg-orange-50 rounded-lg">
                <AlertTriangle className="h-4 w-4 text-orange-600 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-orange-800">
                    Directory Not Empty
                  </p>
                  <p className="text-xs text-orange-700 mt-1">
                    This directory contains files and/or subdirectories. 
                    Choose how you want to proceed:
                  </p>
                </div>
              </div>
              
              <div className="space-y-2">
                <p className="text-sm font-medium">Delete Options:</p>
                <ul className="text-xs text-muted-foreground space-y-1 ml-4">
                  <li>• <strong>Cancel:</strong> Keep the directory and all its contents</li>
                  <li>• <strong>Force Delete:</strong> Delete directory and all contents recursively</li>
                </ul>
              </div>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              Are you sure you want to delete{' '}
              <span className="font-medium">"{itemName}"</span>?
            </p>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isDeleting}
          >
            Cancel
          </Button>
          
          {isDirectory && hasChildren ? (
            <Button
              variant="destructive"
              onClick={() => handleDelete(true)}
              disabled={isDeleting}
              className="bg-red-600 hover:bg-red-700"
            >
              {isDeleting ? 'Deleting...' : 'Force Delete (Recursive)'}
            </Button>
          ) : (
            <Button
              variant="destructive"
              onClick={() => handleDelete(false)}
              disabled={isDeleting}
            >
              {isDeleting ? 'Deleting...' : 'Delete'}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};