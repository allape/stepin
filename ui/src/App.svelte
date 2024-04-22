<script lang="ts">
	import { onMount } from 'svelte';
	import { getCertList } from './api/cert';
	import CertModalButton from './lib/CertModalButton.svelte';

	interface Cert {
		key: string;
		name: string;
		inspection: string;
	}

	let certs: Cert[] = [];

	async function getList() {
		certs = await getCertList();
	}

	onMount(() => {
		getList().then();
	});

	function handleCertChange() {
		getList().then();
	}
</script>

<style lang="scss">
  .wrapper {
    max-width: 1200px;
    margin: auto;
    padding: 10px;
    height: calc(100% - 20px);
    display: flex;
    flex-direction: column;
    justify-content: stretch;
    align-items: stretch;

    .buttons {
      padding: 0 0 10px 0;
      display: flex;
      justify-content: space-between;
    }

    .tableWrapper {
      flex: 1;
      width: 100%;
      overflow: auto;

      .table {
        border-collapse: collapse;
        width: 100%;

				thead {
					position: sticky;
					top: 0;
					left: 0;
					background-color: lightgray;
					color: black;
				}

        th, td {
          border: 1px solid darkgray;
          padding: 3px 5px;
					&.sep {
						background-color: lightgray;
					}
        }
      }
    }
  }
</style>

<div class="wrapper">
	<div class="buttons">
		<div>
			<button on:click={() => window.location.reload()}>Reload</button>
		</div>
		<div>
			<CertModalButton certType="root-ca" on:change={handleCertChange} />
			<CertModalButton certType="intermediate-ca" on:change={handleCertChange} />
			<CertModalButton certType="leaf" on:change={handleCertChange} />
		</div>
	</div>
	<div class="tableWrapper">
		<table class="table">
			<thead>
			<tr>
				<th>Key</th>
				<th>Name</th>
			</tr>
			</thead>
			<tbody>
			{#each certs as cert (cert.key)}
				<tr>
					<td>{cert.key}</td>
					<td>{cert.name}</td>
				</tr>
				<tr>
					<td colspan="2">
						<pre>{cert.inspection}</pre>
					</td>
				</tr>
				<tr>
					<td colspan="2" class="sep"></td>
				</tr>
			{/each}
			</tbody>
		</table>
	</div>
</div>


