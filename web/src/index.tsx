import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import App from './App';
import reportWebVitals from './reportWebVitals';
import './index.css';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
    <BrowserRouter>
      <ConfigProvider
        theme={{
          token: {
            // Demeter — goddess of the harvest.
            // Palette: dust-grey / dry-sage / fern / hunter-green / pine-teal
            colorPrimary: '#588157',
            colorSuccess: '#588157',
            colorWarning: '#a3b18a',
            colorError: '#3a5a40',
            colorInfo: '#588157',
            colorLink: '#588157',
            colorTextBase: '#344e41',
            colorBgBase: '#dad7cd',
            borderRadius: 8,
            fontFamily:
              "'Segoe UI', 'Roboto', 'Helvetica Neue', sans-serif",
          },
          components: {
            Layout: {
              headerBg: '#a3b18a',
              headerColor: '#344e41',
              bodyBg: '#dad7cd',
              siderBg: '#344e41',
            },
            Menu: {
              darkItemBg: '#344e41',
              darkSubMenuItemBg: '#2a3d33',
              darkItemSelectedBg: '#588157',
              darkItemHoverBg: '#3a5a40',
              darkItemColor: '#dad7cd',
              darkItemSelectedColor: '#ffffff',
            },
            Card: {
              colorBgContainer: '#e7e6dd',
            },
          },
        }}
      >
        <App />
      </ConfigProvider>
    </BrowserRouter>
  </React.StrictMode>
);

reportWebVitals();