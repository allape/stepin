import { get } from "@allape/gocrud";
import Crudy, { config } from "@allape/gocrud-react";
import { ICert, ICreateCertBody } from "../model/cert.ts";

export const CertCrudy = new Crudy<ICert>(`${config.SERVER_URL}/cert`);

export function createCert(body: ICreateCertBody): Promise<ICert> {
  return get(`${config.SERVER_URL}/cert/${body._profile}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}
