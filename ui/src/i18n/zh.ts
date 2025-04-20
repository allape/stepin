import { i18n } from "@allape/gocrud-react";
import { TT } from "./en";

const Translation: TT = {
  translation: {
    ...i18n.ZHCN,
    id: "ID",
    name: "证书名称",
    inspection: "证书详情",
    profile: "证书类型",
    unknown: "未知",
    createdAt: "创建时间",
    download: "下载",
    crt: "Crt",
    key: "Key",
    title: "证书管理",
    signNewCertificate: "签发新证书",
    xxxIsRequired: "{{xxx}} 是必填项",
    commonName: "通用名称 / 域名",
    password: "密码",
    lifeSpan: "有效期 (In Year)",
    keyType: "Key 类型",
    parentCA: "上级 CA",
    parentCATips: "子证书或自签名证书时必填",
    parentCaPassword: "上级 CA 密码",
  },
};

export default Translation;
