import { Dashboard } from './components/Dashboard';
import { Toaster } from './components/ui/sonner';

export default function App() {
  return (
    <div className="min-h-screen bg-background">
      <Dashboard />
      <Toaster />
    </div>
  );
}
