import axios from 'axios';
import type {
  MarineArea,
  ProbeResult,
  Regulation,
  SourceInfo,
  Species,
  Technique,
} from './types';

export const apiClient = axios.create({
  baseURL: '/api',
});

export const listSources = async (): Promise<SourceInfo[]> => {
  const { data } = await apiClient.get<SourceInfo[]>('/sources');
  return data;
};

export const runProbe = async (key: string): Promise<ProbeResult> => {
  const { data } = await apiClient.post<ProbeResult>(`/probes/${encodeURIComponent(key)}`);
  return data;
};

export const latestProbe = async (key: string): Promise<ProbeResult | null> => {
  const { data } = await apiClient.get<ProbeResult | null>(
    `/probes/${encodeURIComponent(key)}/latest`,
  );
  return data;
};

export const listSpecies = async (): Promise<Species[]> => {
  const { data } = await apiClient.get<Species[]>('/species');
  return data;
};

export const getSpecies = async (id: number): Promise<Species> => {
  const { data } = await apiClient.get<Species>(`/species/${id}`);
  return data;
};

export type SpeciesPayload = Omit<Species, 'id'>;

export const createSpecies = async (payload: SpeciesPayload): Promise<Species> => {
  const { data } = await apiClient.post<Species>('/species', payload);
  return data;
};

export const updateSpecies = async (
  id: number,
  payload: SpeciesPayload,
): Promise<Species> => {
  const { data } = await apiClient.put<Species>(`/species/${id}`, payload);
  return data;
};

export const listRegulations = async (): Promise<Regulation[]> => {
  const { data } = await apiClient.get<Regulation[]>('/regulations');
  return data;
};

export type RegulationPayload = Omit<Regulation, 'id'>;

export const createRegulation = async (
  payload: RegulationPayload,
): Promise<Regulation> => {
  const { data } = await apiClient.post<Regulation>('/regulations', payload);
  return data;
};

export const updateRegulation = async (
  id: number,
  payload: RegulationPayload,
): Promise<Regulation> => {
  const { data } = await apiClient.put<Regulation>(`/regulations/${id}`, payload);
  return data;
};

export const deleteRegulation = async (id: number): Promise<void> => {
  await apiClient.delete(`/regulations/${id}`);
};

export const listTechniques = async (): Promise<Technique[]> => {
  const { data } = await apiClient.get<Technique[]>('/techniques');
  return data;
};

export type TechniquePayload = Omit<Technique, 'id'>;

export const createTechnique = async (
  payload: TechniquePayload,
): Promise<Technique> => {
  const { data } = await apiClient.post<Technique>('/techniques', payload);
  return data;
};

export const updateTechnique = async (
  id: number,
  payload: TechniquePayload,
): Promise<Technique> => {
  const { data } = await apiClient.put<Technique>(`/techniques/${id}`, payload);
  return data;
};

export const deleteTechnique = async (id: number): Promise<void> => {
  await apiClient.delete(`/techniques/${id}`);
};

export const listMarineAreas = async (): Promise<MarineArea[]> => {
  const { data } = await apiClient.get<MarineArea[]>('/marine-areas');
  return data;
};

export const getMarineArea = async (id: number): Promise<MarineArea> => {
  const { data } = await apiClient.get<MarineArea>(`/marine-areas/${id}`);
  return data;
};
