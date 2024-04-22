/// <reference types="svelte" />
/// <reference types="vite/client" />

type CertType = 'root-ca' | 'intermediate-ca' | 'leaf';

interface Cert {
	key: string;
	name: string;
	inspection: string;
}

interface PutCertBody {
	name: string;
	pass: string;
	years?: number;
	rootCaName?: string;
	rootCaPassword?: string;
}

interface ImportMetaEnv {
	readonly VITE_STEPIN_FE_BASE_URL: string;
}
