import ajax from './http';


export function getCertList(type: CertType | "" = ""): Promise<Cert[]> {
	return ajax<Cert[]>(`/cert?type=${type}`).then(res=> res || []);
}
