<script lang="ts">
	export let open: boolean = false;
</script>

<style lang="scss">
  .pointer {
    cursor: pointer;
    background-color: lightgray;
    padding: 0 4px;
  }
</style>

<div class="wrapper">
	{#if open}
		<slot name="trigger">
			<div class="pointer" on:click={() => open = false}>
				hide
			</div>
		</slot>
		<slot></slot>
	{:else }
		<slot name="closed">
			<div class="pointer" on:click={() => open = true}>
				more
			</div>
		</slot>
	{/if}
</div>
