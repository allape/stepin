// noinspection ExceptionCaughtLocallyJS

import { BASE_URL } from '../config/server';
import { stringify } from '../util/error';

export type Validator<T, D> = (res: D) => Promise<T>;

export async function make<T = unknown, D = unknown>(url: string, validator: Validator<T, D>, options?: RequestInit): Promise<T> {
  try {
    const res = await fetch(url, options);
    // if (res.status != 200) {
    //   throw new Error(res.statusText);
    // }
    return await validator(await res.json());
  } catch (e) {
    const yes = confirm(stringify(e));
    if (yes) {
      return make(url, validator, options);
    }
    throw e;
  }
}

export interface IR<T = unknown> {
  code: string;
  message: string;
  data: T;
}

export default async function ajax<T = unknown>(uri: string, options?: RequestInit): Promise<T> {
  return make<T, IR<T>>(`${BASE_URL}${uri}`,  async (data: IR<T>): Promise<T> => {
    if (data.code !== '0') {
      throw new Error(data.message);
    }
    return data.data;
  }, options);
}
