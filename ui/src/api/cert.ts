import ajax from './http';


export function getCertList(profile: CertProfile | "" = ""): Promise<Cert[]> {
	return ajax<Cert[]>(`/cert?profile=${profile}`).then(res=> res || []);
}
