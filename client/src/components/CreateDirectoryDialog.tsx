import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';

interface CreateDirectoryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentPath: string;
  onDirectoryCreated: () => void;
  onCreateDirectory: (name: string, parentPath: string) => Promise<void>;
}

export const CreateDirectoryDialog: React.FC<CreateDirectoryDialogProps> = ({
  open,
  onOpenChange,
  currentPath,
  onDirectoryCreated,
  onCreateDirectory,
}) => {
  const [directoryName, setDirectoryName] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  const handleCreate = async () => {
    if (!directoryName.trim()) {
      toast.error('Please enter a directory name');
      return;
    }

    setIsCreating(true);
    try {
      await onCreateDirectory(directoryName.trim(), currentPath);
      toast.success(`Directory "${directoryName}" created successfully`);
      setDirectoryName('');
      onOpenChange(false);
      onDirectoryCreated();
    } catch (error) {
      toast.error('Failed to create directory');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create New Directory</DialogTitle>
          <DialogDescription>
            Create a new directory in {currentPath === '/' ? 'root' : currentPath}
          </DialogDescription>
        </DialogHeader>
        
        <div className="py-4">
          <Input
            placeholder="Directory name"
            value={directoryName}
            onChange={(e) => setDirectoryName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                handleCreate();
              }
            }}
            autoFocus
          />
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isCreating}
          >
            Cancel
          </Button>
          <Button
            onClick={handleCreate}
            disabled={isCreating || !directoryName.trim()}
          >
            {isCreating ? 'Creating...' : 'Create Directory'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};