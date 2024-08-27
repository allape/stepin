<script lang="ts">
	export let open: boolean = false;
</script>

<style lang="scss">
  .pointer {
    cursor: pointer;
    padding: 0 4px;
  }
</style>

<div class="wrapper">
	{#if open}
		<div class="pointer" on:click={() => open = false}>
			<slot name="hide">
				Collapse
			</slot>
		</div>
		<slot></slot>
	{:else }
		<div class="pointer" on:click={() => open = true}>
			<slot name="more">
				more
			</slot>
		</div>
	{/if}
</div>
