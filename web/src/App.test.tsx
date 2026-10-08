import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import App from './App';

test('renders hydroponics control panel', () => {
  render(
    <MemoryRouter>
      <ConfigProvider>
        <App />
      </ConfigProvider>
    </MemoryRouter>
  );
  const header = screen.getByText(/Hydroponics Control Panel/i);
  expect(header).toBeInTheDocument();
});