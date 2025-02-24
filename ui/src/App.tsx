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
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
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
        title: t("id"),
        dataIndex: "id",
      },
      {
        title: t("name"),
        dataIndex: "name",
      },
      {
        title: t("inspection"),
        dataIndex: "inspection",
        render: ahelper.EllipsisCell(),
      },
      {
        title: t("profile"),
        dataIndex: "profile",
        render: (v) => {
          const profile = Profiles.find((p) => p.value === v);
          return (
            <Tag color={profile?.color || "gray"}>
              {profile?.label || t("unknown")}
            </Tag>
          );
        },
      },
      {
        title: t("createdAt"),
        dataIndex: "createdAt",
        render: (v) => new Date(v).toLocaleString(),
      },
      {
        title: t("download"),
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
                {t("crt")}
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
                {t("key")}
              </Button>
            </>
          );
        },
      },
    ],
    [t],
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
        title={t("title")}
        extra={
          <>
            <Button type="primary" onClick={openModal}>
              {t("signNewCertificate")}
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
        title={t("signNewCertificate")}
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
            label={t("profile")}
            rules={[
              {
                required: true,
                message: t("xxxIsRequired", { xxx: t("profile") }),
              },
            ]}
          >
            <Select options={Profiles} showSearch optionFilterProp="label" />
          </Form.Item>
          <Form.Item
            name="name"
            label={t("commonName")}
            rules={[
              {
                required: true,
                message: t("xxxIsRequired", { xxx: t("commonName") }),
              },
            ]}
          >
            <Input placeholder="name" allowClear />
          </Form.Item>
          <Form.Item name="pass" label={t("password")}>
            <Input placeholder="password" allowClear />
          </Form.Item>
          <Form.Item
            name="years"
            label={t("lifeSpan")}
            rules={[
              {
                required: true,
                message: t("xxxIsRequired", { xxx: t("lifeSpan") }),
              },
            ]}
          >
            <InputNumber
              placeholder={t("lifeSpan")}
              min={1}
              max={20}
              step={1}
              precision={0}
            />
          </Form.Item>
          <Form.Item
            name="keyType"
            label={t("keyType")}
            rules={[
              {
                required: true,
                message: t("xxxIsRequired", { xxx: t("keyType") }),
              },
            ]}
          >
            <Select options={KeyTypes} showSearch optionFilterProp="label" />
          </Form.Item>
          <Form.Item
            name="parentCaID"
            label={t("parentCA")}
            extra={t("parentCATips")}
          >
            <Select
              options={recordOptions}
              showSearch
              allowClear
              optionFilterProp="label"
              placeholder={t("parentCATips")}
            />
          </Form.Item>
          <Form.Item name="parentCaPassword" label={t("parentCaPassword")}>
            <Input placeholder={t("parentCaPassword")} allowClear />
          </Form.Item>
        </Form>
      </Modal>
    </ThemeProvider>
  );
}
