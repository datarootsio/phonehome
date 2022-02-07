<svelte:head>

  <style>
    @import url("https://fonts.googleapis.com/css2?family=Fira+Sans:wght@100;400;600;900&family=Inconsolata&display=swap");
  </style>

  <link rel="icon" type="image/svg" href="call.svg" />
</svelte:head>

<script>
  import Chart from "svelte-frappe-charts";

  let statsHidden = true;
  let chartData = {};

  let orgRepo = "datarootsio/cheek";
  let serverURL = process.env.SERVER_URL;

  const goButton = () => {
    let [org, repo] = orgRepo.split("/");

    fetch(`${serverURL}/${org}/${repo}/count/daily`)
      .then((response) => response.json())
      .then((data) => {
        let dates = data.data.map((x) => x.date);
        let counts = data.data.map((x) => x.count);

        chartData.labels = dates;
        chartData.datasets = [
          {
            values: counts,
          },
        ];

        statsHidden = false;
      });
  };

  let usageBadgeSrc = () =>
    `https://img.shields.io/endpoint?url=https%3A%2F%2Fapi.phonehome.dev%2F${encodeURIComponent(
      orgRepo
    )}%2Fcount%2Fbadge`;
</script>

<main class="bg-dark-primary min-h-screen">
  <div class="text-white flex items-center justify-end p-2">
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
  <div class="w-10/12 mx-auto">
    <h1
      class="font-mono text-6xl font-bold text-green-basic pt-12 text-center"
    >
      phonehome.dev
    </h1>
    <p class="text-center pt-6 text-purple-basic">
      KISS telemetry for FOSS packages.
    </p>

    <div class="mx-auto flex flex-wrap justify-center gap-4 max-w-fit pt-8">
      <a href="https://api.phonehome.dev/swagger/index.html"><img src="https://img.shields.io/badge/openapi-available-blue?logo=swagger" alt="swagger"></a>
      <a class="inline" href="https://phonehome.dev/coverage.html"><img src="https://img.shields.io/badge/coverage-report-blueviolet?logo=go" alt="coverage"></a>
      <a class="inline" href="https://github.com/datarootsio/phonehome"><img src="https://img.shields.io/badge/docs-README-green?logo=github" alt="readme"></a>

    </div>
    
    <div class="mb-6 pt-24">
      <div class="mb-4 inline">
        <label class="block text-white text-sm mb-2" for="username">
          organisation/repository
        </label>
        <div>
          <input
            bind:value={orgRepo}
            class="shadow appearance-none border rounded w-fill py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline w-64"
            id="username"
            type="text"
            placeholder="datarootsio/phonehome"
          />
          <button
            class="inline bg-blue-500 hover:bg-blue-700 text-white py-2 px-4 rounded"
            on:click={goButton}
          >
            Go
          </button>
          {#if statsHidden}
            <span class="pl-2 text-purple-basic"> 👈 Try me </span>
          {/if}
        </div>
      </div>

      {#if !statsHidden}
        <div class="pt-12">
          <img alt="total-count" src={usageBadgeSrc()} />
          <Chart data={chartData} type="line" />
        </div>
      {/if}
    
    </div>
  </div>
</main>

<style>
</style>
