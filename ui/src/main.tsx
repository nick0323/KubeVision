import { createRoot } from 'react-dom/client';
import App from './App';

/* 全局样式 - 统一导入 */
import './styles/fonts.css';
import './styles/variables.css';
import './styles/layout.css';
import './styles/sidebar.css';
import './styles/table.css';
import './styles/status.css';
import './styles/overview.css';
import './styles/events.css';
import './styles/scrollbar-hide.css';
import './App.css';

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('Failed to find root element');
}

createRoot(rootElement).render(<App />);
