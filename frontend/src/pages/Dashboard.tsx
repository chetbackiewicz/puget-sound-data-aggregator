import {
  Alert,
  Button,
  Card,
  Group,
  SimpleGrid,
  Skeleton,
  Stack,
  Text,
  Title,
} from '@mantine/core';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { listSources } from '../api/client';

const getErrorMessage = (error: unknown) =>
  error instanceof Error ? error.message : 'Unexpected error';

export function Dashboard() {
  const sourcesQuery = useQuery({ queryKey: ['sources'], queryFn: listSources });

  if (sourcesQuery.isLoading) {
    return <Skeleton height={320} radius="md" />;
  }

  if (sourcesQuery.isError) {
    return <Alert color="red">Failed to load dashboard: {getErrorMessage(sourcesQuery.error)}</Alert>;
  }

  const sources = sourcesQuery.data ?? [];
  const dayAgo = Date.now() - 24 * 60 * 60 * 1000;
  const recent = sources.filter((source) => {
    const fetchedAt = source.last_probe?.fetched_at;
    return fetchedAt ? new Date(fetchedAt).getTime() >= dayAgo : false;
  });
  const okRecent = recent.filter((source) => source.last_probe?.ok).length;
  const failedRecent = recent.filter((source) => source.last_probe && !source.last_probe.ok).length;
  const neverProbed = sources.filter((source) => !source.last_probe).length;
  const recentRuns = [...sources]
    .filter((source) => source.last_probe?.fetched_at)
    .sort(
      (a, b) =>
        new Date(b.last_probe!.fetched_at).getTime() - new Date(a.last_probe!.fetched_at).getTime(),
    )
    .slice(0, 5);

  return (
    <Stack gap="lg">
      <Stack gap="xs">
        <Title order={1}>Puget Sound Fishing Data</Title>
        <Text c="dimmed" maw={780}>
          Browse source health, species metadata, regulations, and fishing techniques for Puget
          Sound marine areas.
        </Text>
      </Stack>

      <SimpleGrid cols={{ base: 1, sm: 2, lg: 4 }}>
        {[
          ['Total sources', sources.length],
          ['Sources OK in last 24h', okRecent],
          ['Sources failed in last 24h', failedRecent],
          ['Sources never probed', neverProbed],
        ].map(([label, value]) => (
          <Card key={label} withBorder shadow="sm" radius="md">
            <Text c="dimmed" size="sm">{label}</Text>
            <Title order={2}>{value}</Title>
          </Card>
        ))}
      </SimpleGrid>

      <Card withBorder shadow="sm" radius="md">
        <Stack gap="sm">
          <Group justify="space-between">
            <Title order={3}>Recent probe runs</Title>
            <Button component={Link} to="/probes" variant="light">View probes</Button>
          </Group>
          {recentRuns.length === 0 ? (
            <Text c="dimmed">No probe runs yet.</Text>
          ) : (
            recentRuns.map((source) => (
              <Text key={source.key} component={Link} to="/probes" style={{ textDecoration: 'none' }}>
                {source.last_probe?.ok ? '✅' : '❌'} {source.key} — {source.last_probe?.fetched_at}
              </Text>
            ))
          )}
        </Stack>
      </Card>

      <SimpleGrid cols={{ base: 1, sm: 3 }}>
        {[
          ['Species', '/species', 'Maintain the species catalog.'],
          ['Regulations', '/regulations', 'Review seasons, limits, and emergency rules.'],
          ['Techniques', '/techniques', 'Capture fishing methods and local knowledge.'],
        ].map(([title, to, description]) => (
          <Card key={to} component={Link} to={to} withBorder shadow="sm" radius="md" style={{ textDecoration: 'none' }}>
            <Title order={3}>{title}</Title>
            <Text c="dimmed">{description}</Text>
          </Card>
        ))}
      </SimpleGrid>
    </Stack>
  );
}
