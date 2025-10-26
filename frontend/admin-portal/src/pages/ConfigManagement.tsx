import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Modal,
  Form,
  message,
  Tooltip,
  Switch,
  Popconfirm,
  Card,
  Row,
  Col,
  Statistic,
  Tabs,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  HistoryOutlined,
  LockOutlined,
  UnlockOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useTranslation } from 'react-i18next';
import axios from 'axios';

const { Search } = Input;
const { Option } = Select;
const { TextArea } = Input;
const { TabPane } = Tabs;

interface Config {
  id: string;
  service_name: string;
  config_key: string;
  config_value: string;
  value_type: string;
  environment: string;
  description: string;
  is_encrypted: boolean;
  version: number;
  created_by: string;
  updated_by: string;
  created_at: string;
  updated_at: string;
}

interface ConfigHistory {
  id: string;
  config_id: string;
  old_value: string;
  new_value: string;
  changed_by: string;
  change_reason: string;
  changed_at: string;
}

interface FeatureFlag {
  id: string;
  flag_key: string;
  flag_value: boolean;
  description: string;
  environment: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

const ConfigManagement: React.FC = () => {
  const { t } = useTranslation();
  const [configs, setConfigs] = useState<Config[]>([]);
  const [featureFlags, setFeatureFlags] = useState<FeatureFlag[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [historyModalVisible, setHistoryModalVisible] = useState(false);
  const [editingConfig, setEditingConfig] = useState<Config | null>(null);
  const [configHistory, setConfigHistory] = useState<ConfigHistory[]>([]);
  const [form] = Form.useForm();

  // 筛选条件
  const [filters, setFilters] = useState({
    service_name: 'all',
    environment: 'production',
    search: '',
  });

  // 统计数据
  const [stats, setStats] = useState({
    total: 0,
    encrypted: 0,
    services: 0,
    flags: 0,
  });

  // 加载配置列表
  const loadConfigs = async () => {
    setLoading(true);
    try {
      const params: any = {
        environment: filters.environment,
      };
      if (filters.service_name !== 'all') {
        params.service_name = filters.service_name;
      }

      const response = await axios.get('http://localhost:40010/api/v1/configs', {
        params,
      });

      if (response.data.code === 'SUCCESS') {
        const configList = response.data.data.list || [];
        setConfigs(configList);

        // 计算统计数据
        setStats({
          total: configList.length,
          encrypted: configList.filter((c: Config) => c.is_encrypted).length,
          services: new Set(configList.map((c: Config) => c.service_name)).size,
          flags: featureFlags.length,
        });
      }
    } catch (error) {
      message.error('加载配置失败');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  // 加载功能开关
  const loadFeatureFlags = async () => {
    try {
      const response = await axios.get('http://localhost:40010/api/v1/feature-flags', {
        params: { environment: filters.environment },
      });
      if (response.data.code === 'SUCCESS') {
        setFeatureFlags(response.data.data.list || []);
      }
    } catch (error) {
      console.error('加载功能开关失败:', error);
    }
  };

  useEffect(() => {
    loadConfigs();
    loadFeatureFlags();
  }, [filters]);

  // 新增/编辑配置
  const handleSaveConfig = async (values: any) => {
    try {
      if (editingConfig) {
        // 更新配置
        await axios.put(`http://localhost:40010/api/v1/configs/${editingConfig.id}`, values);
        message.success('配置更新成功');
      } else {
        // 新增配置
        await axios.post('http://localhost:40010/api/v1/configs', values);
        message.success('配置创建成功');
      }
      setModalVisible(false);
      form.resetFields();
      setEditingConfig(null);
      loadConfigs();
    } catch (error) {
      message.error(editingConfig ? '更新失败' : '创建失败');
      console.error(error);
    }
  };

  // 删除配置
  const handleDeleteConfig = async (id: string) => {
    try {
      await axios.delete(`http://localhost:40010/api/v1/configs/${id}`);
      message.success('配置删除成功');
      loadConfigs();
    } catch (error) {
      message.error('删除失败');
      console.error(error);
    }
  };

  // 查看配置历史
  const handleViewHistory = async (configId: string) => {
    try {
      const response = await axios.get(`http://localhost:40010/api/v1/configs/${configId}/history`);
      if (response.data.code === 'SUCCESS') {
        setConfigHistory(response.data.data.list || []);
        setHistoryModalVisible(true);
      }
    } catch (error) {
      message.error('加载历史记录失败');
      console.error(error);
    }
  };

  // 切换功能开关
  const handleToggleFeatureFlag = async (flagKey: string, enabled: boolean) => {
    try {
      await axios.put(`http://localhost:40010/api/v1/feature-flags/${flagKey}`, {
        enabled,
        environment: filters.environment,
      });
      message.success(enabled ? '功能已启用' : '功能已禁用');
      loadFeatureFlags();
    } catch (error) {
      message.error('操作失败');
      console.error(error);
    }
  };

  // 配置表格列
  const columns: ColumnsType<Config> = [
    {
      title: '服务名称',
      dataIndex: 'service_name',
      key: 'service_name',
      width: 150,
      render: (text) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '配置项',
      dataIndex: 'config_key',
      key: 'config_key',
      width: 200,
      render: (text, record) => (
        <Space>
          <span>{text}</span>
          {record.is_encrypted && (
            <Tooltip title="已加密">
              <LockOutlined style={{ color: '#faad14' }} />
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '配置值',
      dataIndex: 'config_value',
      key: 'config_value',
      width: 200,
      ellipsis: true,
      render: (text, record) => {
        if (record.is_encrypted) {
          return <Tag color="orange">****** (已加密)</Tag>;
        }
        return <span>{text}</span>;
      },
    },
    {
      title: '类型',
      dataIndex: 'value_type',
      key: 'value_type',
      width: 100,
      render: (text) => <Tag>{text}</Tag>,
    },
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      width: 100,
      render: (text) => (
        <Tag color={text === 'production' ? 'red' : 'green'}>{text}</Tag>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 80,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: (text) => text ? new Date(text).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => {
                setEditingConfig(record);
                form.setFieldsValue(record);
                setModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="查看历史">
            <Button
              type="link"
              icon={<HistoryOutlined />}
              onClick={() => handleViewHistory(record.id)}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Popconfirm
              title="确定要删除此配置吗？"
              onConfirm={() => handleDeleteConfig(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  // 功能开关表格列
  const flagColumns: ColumnsType<FeatureFlag> = [
    {
      title: '功能标识',
      dataIndex: 'flag_key',
      key: 'flag_key',
      width: 250,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      width: 100,
      render: (text) => <Tag color={text === 'production' ? 'red' : 'green'}>{text}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 100,
      render: (enabled, record) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggleFeatureFlag(record.flag_key, checked)}
        />
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: (text) => new Date(text).toLocaleString(),
    },
  ];

  // 过滤后的配置列表
  const filteredConfigs = configs.filter((config) => {
    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      return (
        config.config_key.toLowerCase().includes(searchLower) ||
        config.config_value.toLowerCase().includes(searchLower) ||
        config.description.toLowerCase().includes(searchLower)
      );
    }
    return true;
  });

  return (
    <div style={{ padding: '24px' }}>
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="配置总数"
              value={stats.total}
              prefix={<SettingOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="加密配置"
              value={stats.encrypted}
              prefix={<LockOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="服务数量"
              value={stats.services}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="功能开关"
              value={stats.flags}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Tabs defaultActiveKey="configs">
        <TabPane tab="配置管理" key="configs">
          <Card>
            {/* 筛选和操作栏 */}
            <Space style={{ marginBottom: 16, width: '100%', justifyContent: 'space-between' }}>
              <Space>
                <Select
                  style={{ width: 200 }}
                  value={filters.service_name}
                  onChange={(value) => setFilters({ ...filters, service_name: value })}
                >
                  <Option value="all">所有服务</Option>
                  <Option value="global">全局配置</Option>
                  <Option value="payment-gateway">payment-gateway</Option>
                  <Option value="order-service">order-service</Option>
                  <Option value="channel-adapter">channel-adapter</Option>
                </Select>

                <Select
                  style={{ width: 150 }}
                  value={filters.environment}
                  onChange={(value) => setFilters({ ...filters, environment: value })}
                >
                  <Option value="production">生产环境</Option>
                  <Option value="development">开发环境</Option>
                  <Option value="staging">预发布环境</Option>
                </Select>

                <Search
                  placeholder="搜索配置项"
                  allowClear
                  style={{ width: 300 }}
                  value={filters.search}
                  onChange={(e) => setFilters({ ...filters, search: e.target.value })}
                  prefix={<SearchOutlined />}
                />

                <Button
                  icon={<ReloadOutlined />}
                  onClick={loadConfigs}
                >
                  刷新
                </Button>
              </Space>

              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  setEditingConfig(null);
                  form.resetFields();
                  setModalVisible(true);
                }}
              >
                新增配置
              </Button>
            </Space>

            {/* 配置表格 */}
            <Table
              columns={columns}
              dataSource={filteredConfigs}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 20,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 条`,
              }}
              scroll={{ x: 1500 }}
            />
          </Card>
        </TabPane>

        <TabPane tab="功能开关" key="flags">
          <Card>
            <Table
              columns={flagColumns}
              dataSource={featureFlags}
              rowKey="id"
              pagination={false}
            />
          </Card>
        </TabPane>
      </Tabs>

      {/* 新增/编辑配置模态框 */}
      <Modal
        title={editingConfig ? '编辑配置' : '新增配置'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          setEditingConfig(null);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        width={700}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveConfig}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="服务名称"
                name="service_name"
                rules={[{ required: true, message: '请输入服务名称' }]}
              >
                <Input placeholder="如: payment-gateway, global" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="环境"
                name="environment"
                rules={[{ required: true, message: '请选择环境' }]}
                initialValue="production"
              >
                <Select>
                  <Option value="production">生产环境</Option>
                  <Option value="development">开发环境</Option>
                  <Option value="staging">预发布环境</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="配置键名"
            name="config_key"
            rules={[{ required: true, message: '请输入配置键名' }]}
          >
            <Input placeholder="如: JWT_SECRET, KAFKA_BROKERS" />
          </Form.Item>

          <Form.Item
            label="配置值"
            name="config_value"
            rules={[{ required: true, message: '请输入配置值' }]}
          >
            <TextArea rows={3} placeholder="配置的具体值" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="值类型"
                name="value_type"
                rules={[{ required: true, message: '请选择值类型' }]}
                initialValue="string"
              >
                <Select>
                  <Option value="string">字符串</Option>
                  <Option value="integer">整数</Option>
                  <Option value="boolean">布尔值</Option>
                  <Option value="json">JSON</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="是否加密"
                name="is_encrypted"
                valuePropName="checked"
                initialValue={false}
              >
                <Switch checkedChildren="是" unCheckedChildren="否" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="描述"
            name="description"
          >
            <TextArea rows={2} placeholder="配置项的说明" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 配置历史模态框 */}
      <Modal
        title="配置变更历史"
        open={historyModalVisible}
        onCancel={() => setHistoryModalVisible(false)}
        footer={null}
        width={900}
      >
        <Table
          dataSource={configHistory}
          rowKey="id"
          pagination={false}
          columns={[
            {
              title: '原值',
              dataIndex: 'old_value',
              key: 'old_value',
              ellipsis: true,
            },
            {
              title: '新值',
              dataIndex: 'new_value',
              key: 'new_value',
              ellipsis: true,
            },
            {
              title: '修改人',
              dataIndex: 'changed_by',
              key: 'changed_by',
              width: 120,
            },
            {
              title: '修改原因',
              dataIndex: 'change_reason',
              key: 'change_reason',
            },
            {
              title: '修改时间',
              dataIndex: 'changed_at',
              key: 'changed_at',
              width: 180,
              render: (text) => new Date(text).toLocaleString(),
            },
          ]}
        />
      </Modal>
    </div>
  );
};

export default ConfigManagement;
