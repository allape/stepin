<script lang="ts">
	import { onMount } from 'svelte';
	import { getCertList } from './api/cert';
	import { BASE_URL } from './config/server';
	import CertModalButton from './lib/CertModalButton.svelte';
	import Click2More from './lib/Click2More.svelte';

	const ColCount = 3;

	let certs: Cert[] = [];

	function sortByName(a: Cert, b: Cert): number {
		return a.name.localeCompare(b.name);
	}

	async function getList() {
		const cs = await getCertList();
		certs = [
			...cs.filter(i => i.profile === 'leaf').sort(sortByName),
			...cs.filter(i => i.profile === 'intermediate-ca').sort(sortByName),
			...cs.filter(i => i.profile === 'root-ca').sort(sortByName)
		];
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

        tr {
          &.root-ca {
            background-color: orangered;
            color: white;
          }

          &.intermediate-ca {
            background-color: gold;
            color: black;
          }

          &.leaf {
            background-color: greenyellow;
            color: black;
          }
        }
        th, td {
          border: 1px solid darkgray;
          padding: 3px 5px;
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
			<CertModalButton profile="root-ca" on:change={handleCertChange} />
			<CertModalButton profile="intermediate-ca" on:change={handleCertChange} />
			<CertModalButton profile="leaf" on:change={handleCertChange} />
		</div>
	</div>
	<div class="tableWrapper">
		<table class="table">
			<thead>
			<tr>
				<th>Key</th>
				<th>Profile</th>
				<th>Name</th>
			</tr>
			</thead>
			<tbody>
			{#each certs as cert (cert.id)}
				<tr class={cert.profile}>
					<td>{cert.name}</td>
					<td>{cert.profile}</td>
					<td>{cert.id}</td>
				</tr>
				<tr class={cert.profile}>
					<td colspan={ColCount}>
						<Click2More>
							<div slot="more">{cert.inspection.split('\n')[0]} ... Click to expand</div>
							<pre>{cert.inspection}</pre>
						</Click2More>
					</td>
				</tr>
				<tr class={cert.profile}>
					<td colspan={ColCount}>
						<a href="{BASE_URL}/cert/crt/{cert.id}" target="_blank">Download Crt</a>
						{#if cert.profile === 'leaf'}
							|
							<a href="{BASE_URL}/cert/key/{cert.id}" target="_blank">Download Key</a>
						{/if}
					</td>
				</tr>
				<tr>
					<td colspan={ColCount}></td>
				</tr>
			{/each}
			</tbody>
		</table>
	</div>
</div>


