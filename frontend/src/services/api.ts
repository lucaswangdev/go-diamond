import axios, { type AxiosInstance, type AxiosError } from 'axios';
import type { Config, ConfigHistory, ApiResponse, ConfigListResponse, CreateConfigRequest, UpdateConfigRequest } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080';

class ApiService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.client.interceptors.request.use((config) => {
      const token = localStorage.getItem('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Config CRUD
  async getConfig(namespace: string, group: string, dataId: string): Promise<Config | null> {
    try {
      const response = await this.client.get<ApiResponse<Config>>(`/api/v1/configs/${namespace}/${group}/${dataId}`);
      return response.data.data || null;
    } catch (error) {
      if ((error as AxiosError).response?.status === 404) {
        return null;
      }
      throw error;
    }
  }

  async createConfig(data: CreateConfigRequest): Promise<{ version: number }> {
    const response = await this.client.post<ApiResponse<{ version: number }>>('/admin/v1/configs', data);
    return response.data.data!;
  }

  async updateConfig(namespace: string, group: string, dataId: string, data: UpdateConfigRequest): Promise<{ version: number }> {
    const response = await this.client.put<ApiResponse<{ version: number }>>(`/admin/v1/configs/${namespace}/${group}/${dataId}`, data);
    return response.data.data!;
  }

  async deleteConfig(namespace: string, group: string, dataId: string, operator: string): Promise<void> {
    await this.client.delete(`/admin/v1/configs/${namespace}/${group}/${dataId}`, {
      data: { operator },
    });
  }

  async listConfigs(namespace: string, group: string, page = 1, pageSize = 20): Promise<ConfigListResponse> {
    const response = await this.client.get<ApiResponse<ConfigListResponse>>('/admin/v1/configs', {
      params: { namespace, group, page, pageSize },
    });
    return response.data.data!;
  }

  async getHistories(namespace: string, group: string, dataId: string, page = 1, pageSize = 20): Promise<{ total: number; list: ConfigHistory[] }> {
    const response = await this.client.get<ApiResponse<{ total: number; list: ConfigHistory[] }>>(
      `/admin/v1/configs/${namespace}/${group}/${dataId}/histories`,
      { params: { page, pageSize } }
    );
    return response.data.data!;
  }

  async rollback(namespace: string, group: string, dataId: string, historyId: number, operator: string): Promise<void> {
    await this.client.post(`/admin/v1/configs/${namespace}/${group}/${dataId}/rollback`, {
      historyId,
      operator,
    });
  }
}

export const apiService = new ApiService();
export default apiService;