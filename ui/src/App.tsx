import { BaseSearchParams, IBaseSearchParams } from "@allape/gocrud";
import { aconfig, ahelper, config, ThemeProvider } from "@allape/gocrud-react";
import { useLoading, useProxy, useToggle } from "@allape/use-loading";
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  message,
  Modal,
  Select,
  Table,
  TableProps,
  Tag,
} from "antd";
import { ReactElement, useCallback, useEffect, useMemo, useState } from "react";
import { CertCrudy, createCert } from "./api/cert.ts";
import {
  ICert,
  ICreateCertBody,
  KeyTypes,
  LV,
  Profile,
  Profiles,
} from "./model/cert.ts";
import styles from "./style.module.scss";

export default function App(): ReactElement {
  const { loading, execute } = useLoading();

  const [visible, _openModal, closeModal] = useToggle(false);

  const [records, recordsRef, setRecords] = useProxy<ICert[]>([]);
  const [recordOptions, setRecordOptions] = useState<LV[]>([]);

  const [form] = Form.useForm<ICreateCertBody>();

  const getList = useCallback(async () => {
    await execute(async () => {
      const records = await CertCrudy.all<IBaseSearchParams>(BaseSearchParams);
      setRecords(records);
      setRecordOptions(records.map((r) => ({ label: r.name, value: r.id })));
    });
  }, [execute, setRecords]);

  useEffect(() => {
    getList().then();
  }, [getList]);

  const handleOk = useCallback(async () => {
    await execute(async () => {
      const data = await form.validateFields();
      const cert = await createCert(data);
      message.success(`Cert ${cert.name} created!`);

      closeModal();

      getList().then();
    });
  }, [closeModal, execute, form, getList]);

  const columns = useMemo<TableProps<ICert>["columns"]>(
    () => [
      {
        title: "ID",
        dataIndex: "id",
      },
      {
        title: "Name",
        dataIndex: "name",
      },
      {
        title: "Inspection",
        dataIndex: "inspection",
        render: ahelper.EllipsisCell(),
      },
      {
        title: "Profile",
        dataIndex: "profile",
        render: (v) => {
          const profile = Profiles.find((p) => p.value === v);
          return (
            <Tag color={profile?.color || "gray"}>
              {profile?.label || "Unknown"}
            </Tag>
          );
        },
      },
      {
        title: "Create Time",
        dataIndex: "createdAt",
        render: (v) => new Date(v).toLocaleString(),
      },
      {
        title: "Actions",
        align: "center",
        fixed: "right",
        render: (_, record) => {
          return (
            <>
              <Button
                size="small"
                type="link"
                href={`${config.SERVER_URL}/cert/crt/${record.id}`}
                target="_blank"
              >
                Crt
              </Button>
              <Button
                disabled={
                  record.profile === "root-ca" ||
                  record.profile === "intermediate-ca"
                }
                size="small"
                type="link"
                href={`${config.SERVER_URL}/cert/key/${record.id}`}
                target="_blank"
              >
                Key
              </Button>
            </>
          );
        },
      },
    ],
    [],
  );

  const openModal = useCallback(() => {
    let profile: Profile;
    let parentCaID: ICreateCertBody["parentCaID"];
    let years: number;

    const certs = [...recordsRef.current].reverse();

    if (recordsRef.current.length === 1) {
      years = 5;
      profile = "intermediate-ca";
      parentCaID = certs.find((i) => i.name.includes("root"))?.id;
    } else if (recordsRef.current.length === 0) {
      profile = "root-ca";
      years = 10;
    } else {
      years = 1;
      profile = "leaf";
      parentCaID = certs.find((i) => i.name.includes("intermedia"))?.id;
    }

    form.setFieldsValue({
      _profile: profile,
      keyType: "EC",
      parentCaID: parentCaID,
      years,
    });
    _openModal();
  }, [_openModal, form, recordsRef]);

  const afterModalClosed = useCallback(() => {
    form.resetFields();
  }, [form]);

  return (
    <ThemeProvider>
      <Card
        className={styles.wrapper}
        title="Certificate Management"
        extra={
          <>
            <Button type="primary" onClick={openModal}>
              Sign New Cert
            </Button>
          </>
        }
      >
        <Table<ICert>
          rowKey="id"
          dataSource={records}
          columns={columns}
          pagination={false}
          scroll={{ x: true }}
        />
      </Card>
      <Modal
        open={visible}
        title="Create a New Certificate"
        closable={!loading}
        maskClosable={!loading}
        onCancel={closeModal}
        cancelButtonProps={{ disabled: loading }}
        onOk={handleOk}
        okButtonProps={{ loading }}
        afterClose={afterModalClosed}
      >
        <Form<ICreateCertBody> {...aconfig.FormLayoutProps} form={form}>
          <Form.Item
            name="_profile"
            label="Profile"
            rules={[{ required: true, message: "Profile is required!" }]}
          >
            <Select options={Profiles} showSearch optionFilterProp="label" />
          </Form.Item>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: "Common Name is required!" }]}
          >
            <Input placeholder="name" allowClear />
          </Form.Item>
          <Form.Item name="pass" label="Password">
            <Input placeholder="password" allowClear />
          </Form.Item>
          <Form.Item
            name="years"
            label="Life Span (In Year)"
            rules={[{ required: true, message: "Life space is required!" }]}
          >
            <InputNumber
              placeholder="years"
              min={1}
              max={20}
              step={1}
              precision={0}
            />
          </Form.Item>
          <Form.Item
            name="keyType"
            label="Key Type"
            rules={[{ required: true, message: "Key type is required!" }]}
          >
            <Select options={KeyTypes} showSearch optionFilterProp="label" />
          </Form.Item>
          <Form.Item
            name="parentCaID"
            label="Parent CA ID"
            extra="Required while signing a leaf or a self-signed cert"
          >
            <Select
              options={recordOptions}
              showSearch
              allowClear
              optionFilterProp="label"
              placeholder="Required while signing a leaf or a self-signed cert"
            />
          </Form.Item>
          <Form.Item name="parentCaPassword" label="Parent CA Password">
            <Input placeholder="password" allowClear />
          </Form.Item>
        </Form>
      </Modal>
    </ThemeProvider>
  );
}
