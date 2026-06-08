import { AppShell, Group, NavLink, Title } from '@mantine/core';
import { IconBook, IconFish, IconGauge, IconHome, IconListCheck } from '@tabler/icons-react';
import { Link, Route, Routes, useLocation } from 'react-router-dom';
import { Dashboard } from './pages/Dashboard';
import { Probes } from './pages/Probes';
import { Regulations } from './pages/Regulations';
import { Species } from './pages/Species';
import { Techniques } from './pages/Techniques';

const links = [
  { label: 'Dashboard', to: '/', icon: IconHome },
  { label: 'Probes', to: '/probes', icon: IconGauge },
  { label: 'Species', to: '/species', icon: IconFish },
  { label: 'Regulations', to: '/regulations', icon: IconListCheck },
  { label: 'Techniques', to: '/techniques', icon: IconBook },
];

function App() {
  const location = useLocation();

  return (
    <AppShell
      header={{ height: 64 }}
      navbar={{ width: 260, breakpoint: 'sm' }}
      padding="md"
    >
      <AppShell.Header px="md">
        <Group h="100%">
          <Title order={3}>Puget Sound Fishing</Title>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        {links.map(({ label, to, icon: Icon }) => (
          <NavLink
            key={to}
            component={Link}
            to={to}
            label={label}
            leftSection={<Icon size={18} stroke={1.5} />}
            active={location.pathname === to}
          />
        ))}
      </AppShell.Navbar>

      <AppShell.Main>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/probes" element={<Probes />} />
          <Route path="/species" element={<Species />} />
          <Route path="/regulations" element={<Regulations />} />
          <Route path="/techniques" element={<Techniques />} />
        </Routes>
      </AppShell.Main>
    </AppShell>
  );
}

export default App;
