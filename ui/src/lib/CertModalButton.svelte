<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { getCertList } from '../api/cert';
	import ajax from '../api/http';
	import { CertTypes } from '../config/cert';
	import Modal from './Modal.svelte';

	const dispatch = createEventDispatcher();

	interface HTMLFormElementX extends HTMLFormElement {
		subjectName: HTMLInputElement;
		pass: HTMLInputElement;
		years: HTMLInputElement;
		rootCaName: HTMLSelectElement;
		rootCaPassword: HTMLInputElement;
	}

	export let certType: CertType = 'leaf';

	let open: boolean = false;
	let formEle: HTMLFormElement | null = null;
	let certs: Cert[] = [];

	$: {
		if (open) {
			const form = formEle as HTMLFormElementX;
			if (form) {
				form.reset();
				switch (certType) {
					case 'root-ca':
						form.subjectName.value = 'root';
						form.years.value = '10';
						break;
					case 'intermediate-ca':
						form.subjectName.value = 'intermediate';
						form.years.value = '10';
						break;
					case 'leaf':
						form.years.value = '1';
						break;
				}
			}
			getCertList().then(list => {
				certs = list;
			});
		}
	}

	async function handleSubmit() {
		const form = formEle as HTMLFormElementX;

		if (!form) {
			return;
		}

		const body: PutCertBody = {
			name: form.subjectName.value,
			pass: form.pass.value,
			rootCaName: form.rootCaName.value,
			rootCaPassword: form.rootCaPassword.value,
			years: Number.parseInt(form.years.value, 10) || 0
		};

		const name = await ajax<string>(`/cert/${certType}`, {
			method: 'PUT',
			body: JSON.stringify(body)
		});
		alert(`${name} saved!`);
		open = false;
		dispatch('change', { name });
	}
</script>

<style lang="scss">
  .wrapper {
    background-color: white;
    padding: 20px;
    max-width: 600px;
    margin: 10px auto 0;

    form {
      table {
        width: 100%;

        input, select {
          width: 100%;
        }
      }
    }

    .buttons {
      position: static;
      bottom: 0;
      left: 0;
    }

    @media (prefers-color-scheme: dark) {
      background-color: black;
    }
  }
</style>

<button on:click={() => open = true}>Add New {CertTypes[certType]} Cert</button>
<Modal bind:open>
	<div class="wrapper">
		<form bind:this={formEle} on:submit|preventDefault>
			<table>
				<tbody>
				<tr>
					<td><label for="SubjectName">Subject Name*:</label></td>
					<td><input id="SubjectName" name="subjectName" type="text"></td>
				</tr>
				<tr>
					<td><label for="Password">Password:</label></td>
					<td><input id="Password" name="pass" type="text"></td>
				</tr>
				<tr>
					<td><label for="years">Life Span(Year):</label></td>
					<td><input id="years" name="years" type="number" min="0" step="1"></td>
				</tr>
				{#if certType !== 'root-ca'}
					<tr>
						<td><label for="RootCAName">Root CA*:</label></td>
						<td>
							<select id="RootCAName" name="rootCaName">
								{#each certs as cert}
									<option value={cert.key}>{cert.key.replace(/^(cert:.+:).+/gi, `$1${cert.name}`)}</option>
								{/each}
							</select>
						</td>
					</tr>
					<tr>
						<td><label for="RootCaPassword">Root CAPassword:</label></td>
						<td><input id="RootCaPassword" name="rootCaPassword" type="text"></td>
					</tr>
				{/if}
				</tbody>
			</table>
		</form>
		<div class="buttons">
			<button on:click={() => open = false}>Close</button>
			<button on:click={handleSubmit}>Submit</button>
		</div>
	</div>
</Modal>
