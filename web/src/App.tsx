import React from 'react';
import { Layout, Menu } from 'antd';
import {
  DashboardOutlined,
  HomeOutlined,
  ExperimentOutlined,
  DeploymentUnitOutlined,
  AimOutlined,
  CameraOutlined,
  ToolOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { Routes, Route, useNavigate, useLocation, Navigate } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import Enclosures from './pages/Enclosures';
import EnclosureDetail from './pages/EnclosureDetail';
import Systems from './pages/Systems';
import SystemDetail from './pages/SystemDetail';
import Flows from './pages/Flows';
import FlowDetail from './pages/FlowDetail';
import Plants from './pages/Plants';
import PlantDetail from './pages/PlantDetail';
import Photos from './pages/Photos';
import Maintenance from './pages/Maintenance';
import Settings from './pages/Settings';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: 'Dashboard' },
  { key: '/enclosures', icon: <HomeOutlined />, label: 'Enclosures' },
  { key: '/systems', icon: <ExperimentOutlined />, label: 'Systems' },
  { key: '/flows', icon: <DeploymentUnitOutlined />, label: 'Flows' },
  { key: '/plants', icon: <AimOutlined />, label: 'Plants' },
  { key: '/photos', icon: <CameraOutlined />, label: 'Photos' },
  { key: '/maintenance', icon: <ToolOutlined />, label: 'Maintenance' },
  { key: '/settings', icon: <SettingOutlined />, label: 'Settings' },
];

const App: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth="0">
        <div
          style={{
            color: '#dad7cd',
            height: 56,
            margin: 8,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 700,
            fontSize: 18,
            letterSpacing: 2,
            textShadow: '0 1px 2px rgba(0,0,0,0.4)',
          }}
        >
          <img
            src={`${process.env.PUBLIC_URL}/logo.svg`}
            alt="Demeter"
            style={{ height: 28, marginRight: 8, filter: 'drop-shadow(0 1px 2px rgba(0,0,0,0.4))' }}
          />
          Demeter
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            fontWeight: 600,
            fontSize: 16,
            letterSpacing: 0.5,
            borderBottom: '2px solid #588157',
          }}
        >
          🌱 Hydroponics Control Panel
        </Header>
        <Content style={{ margin: 24 }}>
          <div style={{ padding: 24, background: '#e7e6dd', minHeight: 360, borderRadius: 8, boxShadow: '0 1px 4px rgba(52, 78, 65, 0.2)' }}>
            <Routes>
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/enclosures" element={<Enclosures />} />
              <Route path="/enclosures/:id" element={<EnclosureDetail />} />
              <Route path="/systems" element={<Systems />} />
              <Route path="/systems/:id" element={<SystemDetail />} />
              <Route path="/flows" element={<Flows />} />
              <Route path="/flows/:id" element={<FlowDetail />} />
              <Route path="/plants" element={<Plants />} />
              <Route path="/plants/:id" element={<PlantDetail />} />
              <Route path="/photos" element={<Photos />} />
              <Route path="/maintenance" element={<Maintenance />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </Routes>
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default App;