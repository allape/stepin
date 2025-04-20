import { IBase } from "@allape/gocrud";
import { IColoredLV } from "@allape/gocrud-react";

export type Profile = "root-ca" | "intermediate-ca" | "leaf" | "self-signed";

export const Profiles: IColoredLV<Profile>[] = [
  {
    label: "RootCA",
    color: "red",
    value: "root-ca",
  },
  {
    label: "IntermediaCA",
    color: "orange",
    value: "intermediate-ca",
  },
  {
    label: "Leaf",
    color: "green",
    value: "leaf",
  },
  // {
  //   label: "Self-Signed",
  //   color: "blue",
  //   value: "self-signed",
  // },
];

export type KeyType = "EC" | "OKP" | "RSA";

export const KeyTypes: IColoredLV<KeyType>[] = [
  {
    label: "EC",
    color: "green",
    value: "EC",
  },
  {
    label: "OKP",
    color: "green",
    value: "OKP",
  },
  {
    label: "RSA",
    color: "orange",
    value: "RSA",
  },
];

export interface ICert extends IBase {
  profile: Profile;
  name: string;
  crt?: string;
  key?: string;
  inspection: string;
}

export interface ICreateCertBody extends Pick<ICert, "name"> {
  _profile: Profile;
  pass?: string;
  years: number;
  keyType: KeyType;
  parentCaID?: number;
  parentCaPassword?: string;
}
