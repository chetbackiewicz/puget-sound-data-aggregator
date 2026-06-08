import {
  Alert,
  Badge,
  Button,
  Group,
  Modal,
  ScrollArea,
  Select,
  Skeleton,
  Stack,
  Table,
  Text,
  Textarea,
  TextInput,
  Title,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import MDEditor from '@uiw/react-md-editor';
import { useMemo, useState } from 'react';
import {
  createTechnique,
  deleteTechnique,
  listSpecies,
  listTechniques,
  updateTechnique,
  type TechniquePayload,
} from '../api/client';
import type { Technique } from '../api/types';

const METHOD_OPTIONS = [
  { value: 'spear', label: 'Spear' },
  { value: 'jig', label: 'Jig' },
  { value: 'troll', label: 'Troll' },
  { value: 'bait', label: 'Bait' },
  { value: 'fly', label: 'Fly' },
  { value: 'other', label: 'Other' },
];

type FormValues = {
  species_id: string;
  title: string;
  method: string;
  recommended_tackle: string;
  recommended_bait: string;
  best_season: string;
  best_time_of_day: string;
  best_conditions: string;
  marine_areas: string;
  body_markdown: string;
  author: string;
  source_attribution: string;
};

const initialValues: FormValues = {
  species_id: '',
  title: '',
  method: '',
  recommended_tackle: '',
  recommended_bait: '',
  best_season: '',
  best_time_of_day: '',
  best_conditions: '',
  marine_areas: '',
  body_markdown: '',
  author: '',
  source_attribution: '',
};

const errMsg = (e: unknown) => (e instanceof Error ? e.message : 'Unexpected error');
const optStr = (v: string) => (v.trim() === '' ? undefined : v.trim());

const toPayload = (v: FormValues): TechniquePayload => ({
  species_id: Number(v.species_id),
  title: v.title.trim(),
  method: optStr(v.method),
  recommended_tackle: optStr(v.recommended_tackle),
  recommended_bait: optStr(v.recommended_bait),
  best_season: optStr(v.best_season),
  best_time_of_day: optStr(v.best_time_of_day),
  best_conditions: optStr(v.best_conditions),
  marine_areas: optStr(v.marine_areas),
  body_markdown: optStr(v.body_markdown),
  author: optStr(v.author),
  source_attribution: optStr(v.source_attribution),
});

const toForm = (t: Technique): FormValues => ({
  species_id: String(t.species_id),
  title: t.title,
  method: t.method ?? '',
  recommended_tackle: t.recommended_tackle ?? '',
  recommended_bait: t.recommended_bait ?? '',
  best_season: t.best_season ?? '',
  best_time_of_day: t.best_time_of_day ?? '',
  best_conditions: t.best_conditions ?? '',
  marine_areas: t.marine_areas ?? '',
  body_markdown: t.body_markdown ?? '',
  author: t.author ?? '',
  source_attribution: t.source_attribution ?? '',
});

export function Techniques() {
  const qc = useQueryClient();
  const [opened, setOpened] = useState(false);
  const [editing, setEditing] = useState<Technique | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<Technique | null>(null);

  const techsQuery = useQuery({ queryKey: ['techniques'], queryFn: listTechniques });
  const speciesQuery = useQuery({ queryKey: ['species'], queryFn: listSpecies });

  const speciesMap = useMemo(
    () => new Map((speciesQuery.data ?? []).map((s) => [s.id, s.common_name])),
    [speciesQuery.data],
  );
  const speciesOptions = useMemo(
    () => (speciesQuery.data ?? []).map((s) => ({ value: String(s.id), label: s.common_name })),
    [speciesQuery.data],
  );

  const form = useForm<FormValues>({
    initialValues,
    validate: {
      species_id: (v) => (v ? null : 'Species is required'),
      title: (v) => (v.trim() ? null : 'Title is required'),
    },
  });

  const saveMutation = useMutation({
    mutationFn: (values: FormValues) => {
      const payload = toPayload(values);
      return editing ? updateTechnique(editing.id, payload) : createTechnique(payload);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['techniques'] });
      notifications.show({ color: 'green', title: 'Technique saved', message: 'Updated.' });
      closeModal();
    },
    onError: (e) => notifications.show({ color: 'red', title: 'Save failed', message: errMsg(e) }),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteTechnique(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['techniques'] });
      notifications.show({ color: 'green', title: 'Deleted', message: 'Technique removed.' });
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

  const openEdit = (t: Technique) => {
    const values = toForm(t);
    setEditing(t);
    form.setValues(values);
    form.resetDirty(values);
    setOpened(true);
  };

  if (techsQuery.isLoading || speciesQuery.isLoading) {
    return <Skeleton height={360} radius="md" />;
  }

  if (techsQuery.isError) {
    return <Alert color="red">Failed to load techniques: {errMsg(techsQuery.error)}</Alert>;
  }

  const techs = techsQuery.data ?? [];

  return (
    <Stack gap="md">
      <Group justify="space-between">
        <Title order={1}>Techniques</Title>
        <Button onClick={openAdd}>Add technique</Button>
      </Group>

      <ScrollArea>
        <Table stickyHeader highlightOnHover withTableBorder>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Species</Table.Th>
              <Table.Th>Title</Table.Th>
              <Table.Th>Method</Table.Th>
              <Table.Th>Best season</Table.Th>
              <Table.Th>Marine areas</Table.Th>
              <Table.Th>Author</Table.Th>
              <Table.Th />
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {techs.map((t) => (
              <Table.Tr key={t.id}>
                <Table.Td>{speciesMap.get(t.species_id) ?? t.species_id}</Table.Td>
                <Table.Td>{t.title}</Table.Td>
                <Table.Td>{t.method ? <Badge variant="light">{t.method}</Badge> : '—'}</Table.Td>
                <Table.Td>{t.best_season ?? '—'}</Table.Td>
                <Table.Td>{t.marine_areas ?? '—'}</Table.Td>
                <Table.Td>{t.author ?? '—'}</Table.Td>
                <Table.Td>
                  <Group gap="xs">
                    <Button size="xs" variant="light" onClick={() => openEdit(t)}>Edit</Button>
                    <Button size="xs" variant="light" color="red" onClick={() => setConfirmDelete(t)}>Delete</Button>
                  </Group>
                </Table.Td>
              </Table.Tr>
            ))}
            {techs.length === 0 && (
              <Table.Tr>
                <Table.Td colSpan={7}>
                  <Text c="dimmed" ta="center">No techniques yet.</Text>
                </Table.Td>
              </Table.Tr>
            )}
          </Table.Tbody>
        </Table>
      </ScrollArea>

      <Modal opened={opened} onClose={closeModal} size="xl" title={editing ? 'Edit technique' : 'Add technique'}>
        <form onSubmit={form.onSubmit((v) => saveMutation.mutate(v))}>
          <Stack>
            <Group grow>
              <Select label="Species" required data={speciesOptions} searchable {...form.getInputProps('species_id')} />
              <Select label="Method" data={METHOD_OPTIONS} clearable {...form.getInputProps('method')} />
            </Group>
            <TextInput label="Title" required {...form.getInputProps('title')} />
            <Group grow>
              <Textarea label="Recommended tackle" minRows={2} {...form.getInputProps('recommended_tackle')} />
              <Textarea label="Recommended bait" minRows={2} {...form.getInputProps('recommended_bait')} />
            </Group>
            <Group grow>
              <TextInput label="Best season" {...form.getInputProps('best_season')} />
              <TextInput label="Best time of day" {...form.getInputProps('best_time_of_day')} />
            </Group>
            <Textarea label="Best conditions" minRows={2} {...form.getInputProps('best_conditions')} />
            <TextInput label="Marine areas" placeholder="MA 7, 8, 9" {...form.getInputProps('marine_areas')} />
            <Stack gap="xs">
              <Text size="sm" fw={500}>Body (markdown)</Text>
              <div data-color-mode="light">
                <MDEditor
                  height={300}
                  value={form.values.body_markdown}
                  onChange={(value) => form.setFieldValue('body_markdown', value ?? '')}
                  preview="edit"
                />
              </div>
            </Stack>
            <Group grow>
              <TextInput label="Author" {...form.getInputProps('author')} />
              <TextInput label="Source attribution" {...form.getInputProps('source_attribution')} />
            </Group>
            <Group justify="flex-end">
              <Button variant="subtle" onClick={closeModal}>Cancel</Button>
              <Button type="submit" loading={saveMutation.isPending}>Save</Button>
            </Group>
          </Stack>
        </form>
      </Modal>

      <Modal opened={confirmDelete !== null} onClose={() => setConfirmDelete(null)} title="Confirm delete">
        <Stack>
          <Text>Delete technique "{confirmDelete?.title}"? This cannot be undone.</Text>
          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setConfirmDelete(null)}>Cancel</Button>
            <Button color="red" loading={deleteMutation.isPending} onClick={() => confirmDelete && deleteMutation.mutate(confirmDelete.id)}>Delete</Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  );
}
