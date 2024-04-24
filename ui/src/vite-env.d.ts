/// <reference types="svelte" />
/// <reference types="vite/client" />

type CertProfile = 'root-ca' | 'intermediate-ca' | 'leaf';

interface Cert {
	id: string;
	profile: CertProfile;
	name: string;
	inspection: string;
}

interface PutCertBody {
	name: string;
	pass: string;
	years?: number;
	rootCaID?: string;
	rootCaPassword?: string;
}

interface ImportMetaEnv {
	readonly VITE_STEPIN_FE_BASE_URL: string;
}
