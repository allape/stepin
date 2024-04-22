<script lang="ts">
	export let open: boolean = false;
</script>

<style lang="scss">
  .modal {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);

    @media (prefers-color-scheme: dark) {
      background-color: rgba(255, 255, 255, 0.5);
    }

		.body {
			top: 0;
			margin: 0 auto;
		}
  }
</style>

{#if open}
	<div class="modal">
		<div class="body">
			<slot></slot>
		</div>
	</div>
{/if}
