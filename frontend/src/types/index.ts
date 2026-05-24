export interface Config {
  id: number;
  namespace: string;
  group: string;
  dataId: string;
  content: string;
  contentMd5: string;
  format: string;
  description: string;
  version: number;
  createdBy: string;
  updatedBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface ConfigHistory {
  id: number;
  configId: number;
  namespace: string;
  group: string;
  dataId: string;
  content: string;
  contentMd5: string;
  version: number;
  opType: string;
  opBy: string;
  createdAt: string;
}

export interface ApiResponse<T> {
  code: number;
  data?: T;
  message?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface CreateConfigRequest {
  namespace: string;
  group: string;
  dataId: string;
  content: string;
  format?: string;
  description?: string;
  operator: string;
}

export interface UpdateConfigRequest {
  content: string;
  format?: string;
  description?: string;
  operator: string;
}

export interface ConfigListResponse {
  total: number;
  list: Config[];
}