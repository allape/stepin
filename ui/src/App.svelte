<script lang="ts">
	import { onMount } from 'svelte';
	import ajax from './api/http';

	let list: string[] = [];

	async function getList() {
		list = (await ajax<string[]>('/leaf')) || [];
	}

	onMount(() => {
		getList().then();
	});
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

        th, td {
          border: 1px solid darkgray;
          padding: 3px 5px;
          white-space: nowrap;
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
			<button on:click={() => window.location.reload()}>Reload</button>
		</div>
	</div>
	<div class="tableWrapper">
		<table class="table">
			<thead>
			<tr>
				<th>Key</th>
				<th>Value</th>
			</tr>
			</thead>
			<tbody>
			{#each list as row}
				<tr>
					<td>{row}</td>
				</tr>
			{/each}
			</tbody>
		</table>
	</div>
</div>
