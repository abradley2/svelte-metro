<script>
	import { writable } from 'svelte/store'

	export let API_URL
	export let WS_URL

	const ws = new WebSocket(WS_URL)

	let geoPosition = null
	let trains = null
	let filteredTrains = null
	let selectedStation = writable(null)
	let stations
	let stationSearchInput
	let stationSearchText = writable("")
	let stationSearchList = []

	fetch(`${API_URL}/stations`)
		.then((e) => e.json())
		.then((data) => {
			stations = data.Stations
			if ($stationSearchText) getStations($stationSearchText)
		})

	ws.onopen = function () {
		ws.onmessage = function (e) {
			try {
				trains = JSON.parse(e.data).Trains
			} catch (err) {
				trains = []
			}
			getTrains(trains)
		}
	}

	stationSearchText.subscribe((function (val) {
		selectedStation.set(null)
		if (!stations) return
		getStations(val)
	}))

	selectedStation.subscribe((val) => {
		if (stationSearchInput && val) {
			stationSearchInput.value = val.Name
			getTrains(val)
		} 
	})

	function getTrains (station) {
		filteredTrains = trains.filter((train) => {
			return train.LocationCode === station.Code
				&& train.Destination !== "Train"
				&& train.Line !== "--"
				&& train.DestinationCode
		})
	}

	function getStations (_searchVal) {
		const searchVal = _searchVal.toUpperCase()

		if (!searchVal) stationSearchList = []

		stationSearchList = stations
			.map((v) => ({...v, searchVal: v.Name.toUpperCase()}))
			.filter((v) => v.searchVal.includes(searchVal))
			.reduce((acc, cur) => {
				const prev = acc.find((v) => v.Name === cur.Name)
				if (prev) return acc
				return acc.concat(cur) 
			}, [])
	}

	function clearStationSearch () {
		stationSearchText.set("")
		stationSearchList = []
	}

	function queryLocation () {
		navigator.geolocation.getCurrentPosition(
			(geo) => {
				geoPosition = geo
				const Lat = geo.coords.latitude
				const Lon = geo.coords.longitude
				
				function dist (station) {
					return Math.acos(
						(
							Math.sin(rads(Lat)) *
							Math.sin(rads(station.Lat))
						) +
						(
							Math.cos(rads(Lat)) *
							Math.cos(rads(station.Lat)) *
							Math.cos(rads(station.Lon - Lon))
						)
					)
				}
				stations = stations.sort((a, b) => {
					return dist(a) < dist(b)
						? -1
						: 1
				})

				stationSearchText.set(stations[0].Name)
				selectedStation.set(stations[0])
				getTrains(stations[0])
			},
			() => {},
			{ enableHighAccuracy: true }
		)

		function rads (deg) {
			return deg * 0.0174533
		}
	}
</script>

<main class="black-70 flex flex-column measure center">
	<div class="pl2 pb2 orange">
		Search by station name
	</div>
	<div
		class="flex justify-between"
	>
		<div>
			<input
			style="height:40px"
			spellcheck="off"
			class="input 
				outline-0 ba b--orange br-pill
				shadow-1 f5 black-70
				w5 pl3 border-box
				hover-bg-orange hover-white orange"
			bind:value={$stationSearchText}
			bind:this={stationSearchInput}
			/>
		{#if $stationSearchText }
			<button
				on:click={clearStationSearch}
				style="height:40px;
					width:40px;
					transform:translateX(-40px);"
				class="button-reset
					ba b--orange
					border-box
					bg-red
					white
					shadow-1
					br-pill"
			>
				X
			</button>
		{/if }
		</div>
		
		{#if !$selectedStation}
		<button
			on:click={queryLocation}
			style="height:40px;"
			class="button-reset bn bg-blue white br-pill mr2
				pointer hover-bg-blue shadow-1 f7 f5-ns
			"
		>
			Use GeoLocation
		</button>
		{/if}
	</div>
	{#if !$selectedStation}
		{#each stationSearchList as station}
			<button
				on:click={() => selectedStation.set(station)}
				class="db
					button-reset br-pill
					bn
					pv2
					bg-orange white ph2
					f4 mt3 pointer
					outline-0
					hover-outline-0
					shadow-1
				"
			>
				{station.Name}
			</button>
		{/each}
	{/if}
	<div class="flex flex-column items-start">
	{#if $selectedStation}
		{#each filteredTrains as train}
			<div
				style="min-height:40px;"
				class="db
					button-reset br-pill
					bn
					pv2
					orange ph2
					f4 mt3
					outline-0
					hover-outline-0
					relative shadow-1
				"
			>
				<span
					style="height:40px;
						width:40px;
						line-height:40px;
						top:calc(50% - 20px)"
					class="dib br-pill
						absolute
						bn
						left-0
						bg-blue
						tc white
						dtc
					"
				>
					{train.Min || "0"}
				</span>
				<span class="pl5 pr3 dib">
					{train.DestinationName}
				</span>
			</div>
		{/each}
	{/if}
	</div>
</main>

<style>

</style>