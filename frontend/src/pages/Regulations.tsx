import {
  Alert,
  Badge,
  Button,
  Group,
  Modal,
  NumberInput,
  ScrollArea,
  Select,
  Skeleton,
  Stack,
  Switch,
  Table,
  Text,
  Textarea,
  TextInput,
  Title,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import {
  createRegulation,
  deleteRegulation,
  listMarineAreas,
  listRegulations,
  listSpecies,
  updateRegulation,
  type RegulationPayload,
} from '../api/client';
import type { Regulation } from '../api/types';

type FormValues = {
  marine_area_id: string;
  species_id: string;
  season_open: string;
  season_close: string;
  daily_limit: number | '';
  size_min_inches: number | '';
  size_max_inches: number | '';
  gear_restrictions: string;
  is_emergency_rule: boolean;
  emergency_rule_expires: string;
  wac_citation: string;
  source_doc: string;
  notes: string;
};

const initialValues: FormValues = {
  marine_area_id: '',
  species_id: '',
  season_open: '',
  season_close: '',
  daily_limit: '',
  size_min_inches: '',
  size_max_inches: '',
  gear_restrictions: '',
  is_emergency_rule: false,
  emergency_rule_expires: '',
  wac_citation: '',
  source_doc: '',
  notes: '',
};

const errMsg = (e: unknown) => (e instanceof Error ? e.message : 'Unexpected error');
const optStr = (v: string) => (v.trim() === '' ? undefined : v.trim());
const optNum = (v: number | '') => (v === '' ? undefined : v);

const toPayload = (v: FormValues): RegulationPayload => ({
  marine_area_id: Number(v.marine_area_id),
  species_id: Number(v.species_id),
  season_open: optStr(v.season_open),
  season_close: optStr(v.season_close),
  daily_limit: optNum(v.daily_limit),
  size_min_inches: optNum(v.size_min_inches),
  size_max_inches: optNum(v.size_max_inches),
  gear_restrictions: optStr(v.gear_restrictions),
  is_emergency_rule: v.is_emergency_rule,
  emergency_rule_expires: v.is_emergency_rule ? optStr(v.emergency_rule_expires) : undefined,
  wac_citation: optStr(v.wac_citation),
  source_doc: optStr(v.source_doc),
  notes: optStr(v.notes),
});

const toForm = (r: Regulation): FormValues => ({
  marine_area_id: String(r.marine_area_id),
  species_id: String(r.species_id),
  season_open: r.season_open ?? '',
  season_close: r.season_close ?? '',
  daily_limit: r.daily_limit ?? '',
  size_min_inches: r.size_min_inches ?? '',
  size_max_inches: r.size_max_inches ?? '',
  gear_restrictions: r.gear_restrictions ?? '',
  is_emergency_rule: r.is_emergency_rule ?? false,
  emergency_rule_expires: r.emergency_rule_expires ?? '',
  wac_citation: r.wac_citation ?? '',
  source_doc: r.source_doc ?? '',
  notes: r.notes ?? '',
});

export function Regulations() {
  const qc = useQueryClient();
  const [opened, setOpened] = useState(false);
  const [editing, setEditing] = useState<Regulation | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<Regulation | null>(null);
  const [filterMA, setFilterMA] = useState<string | null>(null);
  const [filterSp, setFilterSp] = useState<string | null>(null);

  const regsQuery = useQuery({ queryKey: ['regulations'], queryFn: listRegulations });
  const speciesQuery = useQuery({ queryKey: ['species'], queryFn: listSpecies });
  const areasQuery = useQuery({ queryKey: ['marine-areas'], queryFn: listMarineAreas });

  const speciesMap = useMemo(
    () => new Map((speciesQuery.data ?? []).map((s) => [s.id, s.common_name])),
    [speciesQuery.data],
  );
  const areaMap = useMemo(
    () => new Map((areasQuery.data ?? []).map((a) => [a.id, a.name])),
    [areasQuery.data],
  );

  const speciesOptions = useMemo(
    () => (speciesQuery.data ?? []).map((s) => ({ value: String(s.id), label: s.common_name })),
    [speciesQuery.data],
  );
  const areaOptions = useMemo(
    () => (areasQuery.data ?? []).map((a) => ({ value: String(a.id), label: a.name })),
    [areasQuery.data],
  );

  const form = useForm<FormValues>({
    initialValues,
    validate: {
      marine_area_id: (v) => (v ? null : 'Marine area is required'),
      species_id: (v) => (v ? null : 'Species is required'),
    },
  });

  const saveMutation = useMutation({
    mutationFn: (values: FormValues) => {
      const payload = toPayload(values);
      return editing ? updateRegulation(editing.id, payload) : createRegulation(payload);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['regulations'] });
      notifications.show({ color: 'green', title: 'Regulation saved', message: 'Updated.' });
      closeModal();
    },
    onError: (e) => notifications.show({ color: 'red', title: 'Save failed', message: errMsg(e) }),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteRegulation(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['regulations'] });
      notifications.show({ color: 'green', title: 'Deleted', message: 'Regulation removed.' });
      setConfirmDelete(null);
    },
    onError: (e) =>
      notifications.show({ color: 'red', title: 'Delete failed', message: errMsg(e) }),
  });

  const closeModal = () => {
    setOpened(false);
    setEditing(null);
    form.reset();
  };

  const openAdd = () => {
    setEditing(null);
    form.setValues(initialValues);
    form.resetDirty(initialValues);
    setOpened(true);
  };

  const openEdit = (r: Regulation) => {
    const values = toForm(r);
    setEditing(r);
    form.setValues(values);
    form.resetDirty(values);
    setOpened(true);
  };

  const filtered = (regsQuery.data ?? []).filter((r) => {
    if (filterMA && String(r.marine_area_id) !== filterMA) return false;
    if (filterSp && String(r.species_id) !== filterSp) return false;
    return true;
  });

  if (regsQuery.isLoading || speciesQuery.isLoading || areasQuery.isLoading) {
    return <Skeleton height={360} radius="md" />;
  }

  if (regsQuery.isError) {
    return <Alert color="red">Failed to load regulations: {errMsg(regsQuery.error)}</Alert>;
  }

  return (
    <Stack gap="md">
      <Group justify="space-between">
        <Title order={1}>Regulations</Title>
        <Button onClick={openAdd}>Add regulation</Button>
      </Group>

      <Group>
        <Select
          label="Marine area"
          placeholder="All"
          data={areaOptions}
          value={filterMA}
          onChange={setFilterMA}
          clearable
          searchable
        />
        <Select
          label="Species"
          placeholder="All"
          data={speciesOptions}
          value={filterSp}
          onChange={setFilterSp}
          clearable
          searchable
        />
      </Group>

      <ScrollArea>
        <Table stickyHeader highlightOnHover withTableBorder>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Marine area</Table.Th>
              <Table.Th>Species</Table.Th>
              <Table.Th>Season</Table.Th>
              <Table.Th>Daily limit</Table.Th>
              <Table.Th>Size (in)</Table.Th>
              <Table.Th>Status</Table.Th>
              <Table.Th>WAC</Table.Th>
              <Table.Th />
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {filtered.map((r) => (
              <Table.Tr key={r.id}>
                <Table.Td>{areaMap.get(r.marine_area_id) ?? r.marine_area_id}</Table.Td>
                <Table.Td>{speciesMap.get(r.species_id) ?? r.species_id}</Table.Td>
                <Table.Td>
                  {r.season_open ?? '—'} → {r.season_close ?? '—'}
                </Table.Td>
                <Table.Td>{r.daily_limit ?? '—'}</Table.Td>
                <Table.Td>
                  {r.size_min_inches ?? '—'}–{r.size_max_inches ?? '—'}
                </Table.Td>
                <Table.Td>
                  {r.is_emergency_rule ? (
                    <Badge color="red" variant="light">Emergency</Badge>
                  ) : (
                    <Badge variant="light">Regular</Badge>
                  )}
                </Table.Td>
                <Table.Td>
                  <Text size="xs">{r.wac_citation ?? '—'}</Text>
                </Table.Td>
                <Table.Td>
                  <Group gap="xs">
                    <Button size="xs" variant="light" onClick={() => openEdit(r)}>Edit</Button>
                    <Button size="xs" variant="light" color="red" onClick={() => setConfirmDelete(r)}>Delete</Button>
                  </Group>
                </Table.Td>
              </Table.Tr>
            ))}
            {filtered.length === 0 && (
              <Table.Tr>
                <Table.Td colSpan={8}>
                  <Text c="dimmed" ta="center">No regulations match filters.</Text>
                </Table.Td>
              </Table.Tr>
            )}
          </Table.Tbody>
        </Table>
      </ScrollArea>

      <Modal
        opened={opened}
        onClose={closeModal}
        size="lg"
        title={editing ? 'Edit regulation' : 'Add regulation'}
      >
        <form onSubmit={form.onSubmit((v) => saveMutation.mutate(v))}>
          <Stack>
            <Group grow>
              <Select label="Marine area" required data={areaOptions} searchable {...form.getInputProps('marine_area_id')} />
              <Select label="Species" required data={speciesOptions} searchable {...form.getInputProps('species_id')} />
            </Group>
            <Group grow>
              <TextInput label="Season open" placeholder="2025-07-01" {...form.getInputProps('season_open')} />
              <TextInput label="Season close" placeholder="2025-09-30" {...form.getInputProps('season_close')} />
            </Group>
            <Group grow>
              <NumberInput label="Daily limit" min={0} {...form.getInputProps('daily_limit')} />
              <NumberInput label="Min size (in)" min={0} step={0.5} {...form.getInputProps('size_min_inches')} />
              <NumberInput label="Max size (in)" min={0} step={0.5} {...form.getInputProps('size_max_inches')} />
            </Group>
            <Textarea label="Gear restrictions" minRows={2} {...form.getInputProps('gear_restrictions')} />
            <Switch label="Emergency rule" {...form.getInputProps('is_emergency_rule', { type: 'checkbox' })} />
            {form.values.is_emergency_rule && (
              <TextInput label="Emergency rule expires" placeholder="2025-12-31" {...form.getInputProps('emergency_rule_expires')} />
            )}
            <Group grow>
              <TextInput label="WAC citation" {...form.getInputProps('wac_citation')} />
              <TextInput label="Source doc" {...form.getInputProps('source_doc')} />
            </Group>
            <Textarea label="Notes" minRows={2} {...form.getInputProps('notes')} />
            <Group justify="flex-end">
              <Button variant="subtle" onClick={closeModal}>Cancel</Button>
              <Button type="submit" loading={saveMutation.isPending}>Save</Button>
            </Group>
          </Stack>
        </form>
      </Modal>

      <Modal opened={confirmDelete !== null} onClose={() => setConfirmDelete(null)} title="Confirm delete">
        <Stack>
          <Text>
            Delete regulation #{confirmDelete?.id} for {confirmDelete ? speciesMap.get(confirmDelete.species_id) : ''}? This cannot be undone.
          </Text>
          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setConfirmDelete(null)}>Cancel</Button>
            <Button color="red" loading={deleteMutation.isPending} onClick={() => confirmDelete && deleteMutation.mutate(confirmDelete.id)}>Delete</Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  );
}
