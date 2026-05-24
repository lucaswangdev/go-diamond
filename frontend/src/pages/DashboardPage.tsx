import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import apiService from '../services/api';
import type { Config, ConfigHistory } from '../types';

export default function DashboardPage() {
  const { user, logout } = useAuth();
  const [configs, setConfigs] = useState<Config[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [selectedConfig, setSelectedConfig] = useState<Config | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [isEditing, setIsEditing] = useState(false);

  // Form state
  const [namespace, setNamespace] = useState('default');
  const [group, setGroup] = useState('DEFAULT_GROUP');
  const [dataId, setDataId] = useState('');
  const [content, setContent] = useState('');
  const [format, setFormat] = useState('json');
  const [description, setDescription] = useState('');

  // History state
  const [showHistory, setShowHistory] = useState(false);
  const [histories, setHistories] = useState<ConfigHistory[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);

  const loadConfigs = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await apiService.listConfigs(namespace, group, 1, 100);
      setConfigs(data.list);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfigs();
  }, [namespace, group]);

  const handleCreate = async () => {
    if (!dataId || !content) {
      setError('DataID and Content are required');
      return;
    }

    try {
      await apiService.createConfig({
        namespace,
        group,
        dataId,
        content,
        format,
        description,
        operator: user?.name || 'unknown',
      });
      setIsCreating(false);
      resetForm();
      loadConfigs();
    } catch (err) {
      setError((err as Error).message);
    }
  };

  const handleUpdate = async () => {
    if (!selectedConfig) return;

    try {
      await apiService.updateConfig(selectedConfig.namespace, selectedConfig.group, selectedConfig.dataId, {
        content,
        format,
        description,
        operator: user?.name || 'unknown',
      });
      setIsEditing(false);
      setSelectedConfig(null);
      resetForm();
      loadConfigs();
    } catch (err) {
      setError((err as Error).message);
    }
  };

  const handleDelete = async (config: Config) => {
    if (!confirm(`Delete config ${config.dataId}?`)) return;

    try {
      await apiService.deleteConfig(config.namespace, config.group, config.dataId, user?.name || 'unknown');
      loadConfigs();
    } catch (err) {
      setError((err as Error).message);
    }
  };

  const handleRollback = async (historyId: number) => {
    if (!selectedConfig) return;
    if (!confirm('Rollback to this version?')) return;

    try {
      await apiService.rollback(selectedConfig.namespace, selectedConfig.group, selectedConfig.dataId, historyId, user?.name || 'unknown');
      setShowHistory(false);
      loadConfigs();
    } catch (err) {
      setError((err as Error).message);
    }
  };

  const loadHistories = async () => {
    if (!selectedConfig) return;
    setHistoryLoading(true);
    try {
      const data = await apiService.getHistories(selectedConfig.namespace, selectedConfig.group, selectedConfig.dataId);
      setHistories(data.list);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setHistoryLoading(false);
    }
  };

  const openEdit = (config: Config) => {
    setSelectedConfig(config);
    setNamespace(config.namespace);
    setGroup(config.group);
    setDataId(config.dataId);
    setContent(config.content);
    setFormat(config.format);
    setDescription(config.description);
    setIsEditing(true);
    setIsCreating(false);
  };

  const openHistory = async (config: Config) => {
    setSelectedConfig(config);
    setShowHistory(true);
    await loadHistories();
  };

  const resetForm = () => {
    setDataId('');
    setContent('');
    setFormat('json');
    setDescription('');
    setError('');
  };

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-900">go-diamond</h1>
          <div className="flex items-center gap-4">
            <span className="text-gray-600">{user?.email}</span>
            <button
              onClick={logout}
              className="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-md"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6">
        {/* Filters */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <div className="flex gap-4 items-end flex-wrap">
            <div>
              <label className="block text-sm font-medium text-gray-700">Namespace</label>
              <input
                type="text"
                value={namespace}
                onChange={(e) => setNamespace(e.target.value)}
                className="mt-1 block w-40 px-3 py-2 border border-gray-300 rounded-md"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Group</label>
              <input
                type="text"
                value={group}
                onChange={(e) => setGroup(e.target.value)}
                className="mt-1 block w-40 px-3 py-2 border border-gray-300 rounded-md"
              />
            </div>
            <button
              onClick={loadConfigs}
              className="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700"
            >
              Search
            </button>
            <button
              onClick={() => { setIsCreating(true); setIsEditing(false); resetForm(); }}
              className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
            >
              + New Config
            </button>
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="bg-red-50 text-red-500 p-4 rounded-md mb-4">{error}</div>
        )}

        {/* Config List */}
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Namespace</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Group</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">DataID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Format</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Version</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Updated</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {loading ? (
                <tr>
                  <td colSpan={7} className="px-6 py-4 text-center text-gray-500">Loading...</td>
                </tr>
              ) : configs.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-6 py-4 text-center text-gray-500">No configs found</td>
                </tr>
              ) : (
                configs.map((config) => (
                  <tr key={`${config.namespace}-${config.group}-${config.dataId}`}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{config.namespace}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{config.group}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-mono">{config.dataId}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{config.format}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">v{config.version}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(config.updatedAt).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <div className="flex gap-2">
                        <button
                          onClick={() => openEdit(config)}
                          className="text-purple-600 hover:text-purple-800"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => openHistory(config)}
                          className="text-blue-600 hover:text-blue-800"
                        >
                          History
                        </button>
                        <button
                          onClick={() => handleDelete(config)}
                          className="text-red-600 hover:text-red-800"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </main>

      {/* Create/Edit Modal */}
      {(isCreating || isEditing) && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-auto">
            <div className="p-6 border-b">
              <h2 className="text-xl font-semibold">{isCreating ? 'Create Config' : 'Edit Config'}</h2>
            </div>
            <div className="p-6 space-y-4">
              {isCreating && (
                <>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700">Namespace</label>
                      <input
                        type="text"
                        value={namespace}
                        onChange={(e) => setNamespace(e.target.value)}
                        className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700">Group</label>
                      <input
                        type="text"
                        value={group}
                        onChange={(e) => setGroup(e.target.value)}
                        className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">DataID</label>
                    <input
                      type="text"
                      value={dataId}
                      onChange={(e) => setDataId(e.target.value)}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                      placeholder="app.json"
                    />
                  </div>
                </>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700">Content</label>
                <textarea
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  rows={10}
                  className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md font-mono text-sm"
                  placeholder='{"key": "value"}'
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Format</label>
                  <select
                    value={format}
                    onChange={(e) => setFormat(e.target.value)}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                  >
                    <option value="json">JSON</option>
                    <option value="text">Text</option>
                    <option value="yaml">YAML</option>
                    <option value="xml">XML</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  <input
                    type="text"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                  />
                </div>
              </div>
            </div>
            <div className="p-6 border-t flex justify-end gap-4">
              <button
                onClick={() => { setIsCreating(false); setIsEditing(false); setSelectedConfig(null); resetForm(); }}
                className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-md"
              >
                Cancel
              </button>
              <button
                onClick={isCreating ? handleCreate : handleUpdate}
                className="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700"
              >
                {isCreating ? 'Create' : 'Update'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* History Modal */}
      {showHistory && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-auto">
            <div className="p-6 border-b">
              <h2 className="text-xl font-semibold">
                History: {selectedConfig?.namespace}/{selectedConfig?.group}/{selectedConfig?.dataId}
              </h2>
            </div>
            <div className="p-6">
              {historyLoading ? (
                <p className="text-center text-gray-500">Loading...</p>
              ) : histories.length === 0 ? (
                <p className="text-center text-gray-500">No history found</p>
              ) : (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Version</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Op</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">By</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Time</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Content</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Action</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {histories.map((h) => (
                      <tr key={h.id}>
                        <td className="px-4 py-3 text-sm">v{h.version}</td>
                        <td className="px-4 py-3 text-sm">
                          <span className={`px-2 py-1 text-xs rounded ${
                            h.opType === 'create' ? 'bg-green-100 text-green-800' :
                            h.opType === 'update' ? 'bg-blue-100 text-blue-800' :
                            'bg-red-100 text-red-800'
                          }`}>
                            {h.opType}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-sm">{h.opBy}</td>
                        <td className="px-4 py-3 text-sm text-gray-500">
                          {new Date(h.createdAt).toLocaleString()}
                        </td>
                        <td className="px-4 py-3 text-sm font-mono max-w-xs truncate">
                          {h.content}
                        </td>
                        <td className="px-4 py-3 text-sm">
                          <button
                            onClick={() => handleRollback(h.id)}
                            className="text-purple-600 hover:text-purple-800"
                          >
                            Rollback
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            <div className="p-6 border-t">
              <button
                onClick={() => { setShowHistory(false); setSelectedConfig(null); }}
                className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-md"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}