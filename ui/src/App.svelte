<script>
  import Chart from "svelte-frappe-charts";

  let data = {
    labels: ["Sun", "Mon", "Tues", "Wed", "Thurs", "Fri", "Sat"],
    datasets: [
      {
        values: [10, 12, 3, 9, 8, 15, 9],
      },
    ],
  };
</script>

<svelte:head>
  <style>
    @import url("https://fonts.googleapis.com/css2?family=Fira+Sans:wght@100;400;600;900&family=Inconsolata&display=swap");
  </style>
</svelte:head>

<main class="bg-dark-primary w-full h-screen">
  <div class="text-white flex items-center justify-end w-full p-2">
    <div class="right-0">
      <span class="pr-1 text-xs">by</span>
      <a href="https://dataroots.io"
        ><img
          class="w-24 inline"
          src="/dataroots-logo-white.svg"
          alt="dataroots"
        /></a
      >
    </div>
  </div>
  <div class="container mx-auto">
    <h1
      class="auto-mx font-mono text-6xl font-bold text-green-basic pt-12 text-center"
    >
      phonehome.dev
    </h1>
    <p class="pt-6 text-center text-purple-basic">
      KISS telemetry for FOSS packages.
    </p>
    <div class="mb-6 pt-12">
      <div class="mb-4 inline">
        <label class="block text-white text-sm mb-2" for="username">
          organisation/repository
        </label>
        <div>
          <input
            class="shadow appearance-none border rounded w-fill py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline w-64"
            id="username"
            type="text"
            placeholder="datarootsio/phonehome"
          />
          <button
            class="inline bg-blue-500 hover:bg-blue-700 text-white py-2 px-4 rounded"
          >
            Button
          </button>
        </div>
      </div>

      <div class="pt-12">
        <Chart {data} type="line" />
      </div>
    </div>
  </div>
</main>

<style>
</style>
