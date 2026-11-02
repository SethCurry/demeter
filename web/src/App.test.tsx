import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import App from './App';

beforeEach(() => {
  global.fetch = jest.fn().mockResolvedValue({
    ok: true,
    status: 200,
    statusText: 'OK',
    json: async () => [],
  }) as unknown as typeof fetch;
});

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders hydroponics control panel', async () => {
  render(
    <MemoryRouter>
      <ConfigProvider>
        <App />
      </ConfigProvider>
    </MemoryRouter>
  );
  const header = await screen.findByText(/Hydroponics Control Panel/i);
  expect(header).toBeInTheDocument();
});
