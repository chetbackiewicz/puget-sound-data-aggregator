import {
  Alert,
  Badge,
  Button,
  Card,
  Code,
  Grid,
  Group,
  ScrollArea,
  Skeleton,
  Stack,
  Text,
  Title,
} from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { listSources, runProbe } from '../api/client';
import type { ProbeResult } from '../api/types';

const getErrorMessage = (error: unknown) =>
  error instanceof Error ? error.message : 'Unexpected error';

export function Probes() {
  const queryClient = useQueryClient();
  const [probeResults, setProbeResults] = useState<Record<string, ProbeResult>>({});
  const sourcesQuery = useQuery({ queryKey: ['sources'], queryFn: listSources });

  const probeMutation = useMutation({
    mutationFn: runProbe,
    onSuccess: (result) => {
      setProbeResults((current) => ({ ...current, [result.source_key]: result }));
      queryClient.invalidateQueries({ queryKey: ['sources'] });
      notifications.show({ color: 'green', title: 'Probe complete', message: result.source_key });
    },
    onError: (error) => {
      notifications.show({ color: 'red', title: 'Probe failed', message: getErrorMessage(error) });
    },
  });

  if (sourcesQuery.isLoading) {
    return <Skeleton height={360} radius="md" />;
  }

  if (sourcesQuery.isError) {
    return <Alert color="red">Failed to load sources: {getErrorMessage(sourcesQuery.error)}</Alert>;
  }

  return (
    <Stack gap="md">
      <Title order={1}>Source probes</Title>
      <Grid>
        {sourcesQuery.data?.map((source) => {
          const lastProbe = source.last_probe;
          const expandedResult = probeResults[source.key];
          const isRunning = probeMutation.isPending && probeMutation.variables === source.key;

          return (
            <Grid.Col key={source.key} span={{ base: 12, md: 6, xl: 4 }}>
              <Card withBorder shadow="sm" radius="md" h="100%">
                <Stack gap="sm">
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{source.key}</Title>
                    {source.auth_required && <Badge color="orange">auth required</Badge>}
                  </Group>
                  <Text c="dimmed">{source.description}</Text>
                  <Group gap="xs">
                    <Text fw={600}>Status:</Text>
                    <Text c={lastProbe ? (lastProbe.ok ? 'green' : 'red') : 'gray'}>
                      {lastProbe ? (lastProbe.ok ? '✅' : '❌') : '—'}
                    </Text>
                  </Group>
                  <Text size="sm">Fetched: {lastProbe?.fetched_at ?? 'Never'}</Text>
                  <Text size="sm">HTTP: {lastProbe?.http_status ?? '—'}</Text>
                  <Text size="sm">Duration: {lastProbe?.duration_ms ?? '—'} ms</Text>
                  <Button
                    onClick={() => probeMutation.mutate(source.key)}
                    loading={isRunning}
                    disabled={probeMutation.isPending && !isRunning}
                  >
                    Run probe
                  </Button>
                  {expandedResult && (
                    <Stack gap="xs">
                      <Text fw={600}>Parsed summary</Text>
                      <ScrollArea.Autosize mah={400} type="auto">
                        <Code block>{JSON.stringify(expandedResult.parsed_summary ?? null, null, 2)}</Code>
                      </ScrollArea.Autosize>
                      <Text fw={600}>Raw response snippet</Text>
                      <ScrollArea.Autosize mah={400} type="auto">
                        <Code block>{expandedResult.raw_response_snippet ?? 'No response snippet'}</Code>
                      </ScrollArea.Autosize>
                    </Stack>
                  )}
                </Stack>
              </Card>
            </Grid.Col>
          );
        })}
      </Grid>
    </Stack>
  );
}
