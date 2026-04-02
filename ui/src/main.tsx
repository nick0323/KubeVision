import { createRoot } from 'react-dom/client';
import './index.css';
import './components/ScrollbarHide.css';
import App from './App';

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('Failed to find root element');
}

// StrictMode 会导致组件挂载两次，与 WebSocket 重连逻辑冲突
// 开发环境可以启用，但需要改进重连逻辑
createRoot(rootElement).render(<App />);
