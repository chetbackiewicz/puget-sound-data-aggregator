import {
  Alert,
  Badge,
  Button,
  Group,
  Modal,
  NumberInput,
  ScrollArea,
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
import { useState } from 'react';
import { createSpecies, listSpecies, updateSpecies, type SpeciesPayload } from '../api/client';
import type { Species as SpeciesType } from '../api/types';

type SpeciesFormValues = {
  scientific_name: string;
  common_name: string;
  worms_aphia_id: number | '';
  gbif_taxon_key: number | '';
  inat_taxon_id: number | '';
  wikidata_qid: string;
  fishbase_speccode: number | '';
  spear_legal: boolean;
  notes: string;
};

const initialValues: SpeciesFormValues = {
  scientific_name: '',
  common_name: '',
  worms_aphia_id: '',
  gbif_taxon_key: '',
  inat_taxon_id: '',
  wikidata_qid: '',
  fishbase_speccode: '',
  spear_legal: false,
  notes: '',
};

const getErrorMessage = (error: unknown) =>
  error instanceof Error ? error.message : 'Unexpected error';

const optionalNumber = (value: number | '') => (value === '' ? undefined : value);
const optionalString = (value: string) => (value.trim() === '' ? undefined : value.trim());

const toPayload = (values: SpeciesFormValues): SpeciesPayload => ({
  scientific_name: values.scientific_name.trim(),
  common_name: values.common_name.trim(),
  worms_aphia_id: optionalNumber(values.worms_aphia_id),
  gbif_taxon_key: optionalNumber(values.gbif_taxon_key),
  inat_taxon_id: optionalNumber(values.inat_taxon_id),
  wikidata_qid: optionalString(values.wikidata_qid),
  fishbase_speccode: optionalNumber(values.fishbase_speccode),
  spear_legal: values.spear_legal,
  notes: optionalString(values.notes),
});

const toFormValues = (species: SpeciesType): SpeciesFormValues => ({
  scientific_name: species.scientific_name,
  common_name: species.common_name,
  worms_aphia_id: species.worms_aphia_id ?? '',
  gbif_taxon_key: species.gbif_taxon_key ?? '',
  inat_taxon_id: species.inat_taxon_id ?? '',
  wikidata_qid: species.wikidata_qid ?? '',
  fishbase_speccode: species.fishbase_speccode ?? '',
  spear_legal: species.spear_legal ?? false,
  notes: species.notes ?? '',
});

function IdBadge({ label, value }: { label: string; value?: number | string }) {
  return value ? <Badge size="sm" variant="light">{label}: {value}</Badge> : null;
}

export function Species() {
  const queryClient = useQueryClient();
  const [opened, setOpened] = useState(false);
  const [editing, setEditing] = useState<SpeciesType | null>(null);
  const speciesQuery = useQuery({ queryKey: ['species'], queryFn: listSpecies });
  const form = useForm<SpeciesFormValues>({
    initialValues,
    validate: {
      scientific_name: (value) => (value.trim() ? null : 'Scientific name is required'),
      common_name: (value) => (value.trim() ? null : 'Common name is required'),
    },
  });

  const saveMutation = useMutation({
    mutationFn: (values: SpeciesFormValues) => {
      const payload = toPayload(values);
      return editing ? updateSpecies(editing.id, payload) : createSpecies(payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['species'] });
      notifications.show({ color: 'green', title: 'Species saved', message: 'Catalog updated.' });
      closeModal();
    },
    onError: (error) => {
      notifications.show({ color: 'red', title: 'Save failed', message: getErrorMessage(error) });
    },
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

  const openEdit = (species: SpeciesType) => {
    const values = toFormValues(species);
    setEditing(species);
    form.setValues(values);
    form.resetDirty(values);
    setOpened(true);
  };

  if (speciesQuery.isLoading) {
    return <Skeleton height={360} radius="md" />;
  }

  if (speciesQuery.isError) {
    return <Alert color="red">Failed to load species: {getErrorMessage(speciesQuery.error)}</Alert>;
  }

  return (
    <Stack gap="md">
      <Group justify="space-between">
        <Title order={1}>Species</Title>
        <Button onClick={openAdd}>Add species</Button>
      </Group>

      <ScrollArea>
        <Table stickyHeader highlightOnHover withTableBorder>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Common name</Table.Th>
              <Table.Th>Scientific name</Table.Th>
              <Table.Th>Spear legal</Table.Th>
              <Table.Th>External IDs</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {speciesQuery.data?.map((species) => (
              <Table.Tr key={species.id} onClick={() => openEdit(species)} style={{ cursor: 'pointer' }}>
                <Table.Td>{species.common_name}</Table.Td>
                <Table.Td><Text fs="italic">{species.scientific_name}</Text></Table.Td>
                <Table.Td>{species.spear_legal ? '✓' : '✗'}</Table.Td>
                <Table.Td>
                  <Group gap={4}>
                    <IdBadge label="WoRMS" value={species.worms_aphia_id} />
                    <IdBadge label="GBIF" value={species.gbif_taxon_key} />
                    <IdBadge label="iNat" value={species.inat_taxon_id} />
                    <IdBadge label="Wikidata" value={species.wikidata_qid} />
                  </Group>
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      </ScrollArea>

      <Modal opened={opened} onClose={closeModal} title={editing ? 'Edit species' : 'Add species'} size="lg">
        <form onSubmit={form.onSubmit((values) => saveMutation.mutate(values))}>
          <Stack>
            <TextInput label="Scientific name" required {...form.getInputProps('scientific_name')} />
            <TextInput label="Common name" required {...form.getInputProps('common_name')} />
            <Group grow>
              <NumberInput label="WoRMS Aphia ID" min={0} {...form.getInputProps('worms_aphia_id')} />
              <NumberInput label="GBIF taxon key" min={0} {...form.getInputProps('gbif_taxon_key')} />
            </Group>
            <Group grow>
              <NumberInput label="iNaturalist taxon ID" min={0} {...form.getInputProps('inat_taxon_id')} />
              <NumberInput label="FishBase spec code" min={0} {...form.getInputProps('fishbase_speccode')} />
            </Group>
            <TextInput label="Wikidata QID" {...form.getInputProps('wikidata_qid')} />
            <Switch label="Spear legal" {...form.getInputProps('spear_legal', { type: 'checkbox' })} />
            <Textarea label="Notes" minRows={3} {...form.getInputProps('notes')} />
            <Group justify="flex-end">
              <Button variant="subtle" onClick={closeModal}>Cancel</Button>
              <Button type="submit" loading={saveMutation.isPending}>Save</Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </Stack>
  );
}
