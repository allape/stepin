import { i18n } from "@allape/gocrud-react";

const Translation = {
  ...i18n,
  id: "ID",
  name: "Name",
  inspection: "Inspection",
  profile: "Profile",
  unknown: "Unknown",
  createdAt: "Create Time",
  download: "Download",
  crt: "Crt",
  key: "Key",
  title: "Certificate Management",
  signNewCertificate: "Sign New Certificate",
  xxxIsRequired: "{{xxx}} is required",
  commonName: "Common Name",
  password: "Password",
  lifeSpan: "Life Span (In Year)",
  keyType: "Key Type",
  parentCA: "Parent CA",
  parentCATips: "Required while signing a leaf or a self-signed cert",
  parentCaPassword: "Parent CA Password",
};

export type TT = typeof Translation;

export default Translation;
