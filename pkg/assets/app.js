function fetch_ok(path, post_params = undefined) {

	let params = {};
	if(post_params !== undefined) {
		let d = new FormData();
		for(let [key, value] of Object.entries(post_params)) {
			d.append(key, value);
		}
		params.method = "POST";
		params.body = d;
	}

	return fetch(path, params).then((resp) => {
		if(!resp.ok) {
			throw resp.statusText;
		}
		return resp;
	});
}

document.addEventListener("DOMContentLoaded", async () => {

	let resp = await fetch_ok("/commands");
	let commands = await resp.json();

	let list = document.querySelector("#list");
	let item_template = list.querySelector(":scope > template");

	for(let cmd of commands) {

		let elem = item_template.content.firstElementChild.cloneNode(true);
		list.appendChild(elem);

		elem.dataset.name = cmd.name;
		elem.querySelector(".name").textContent = cmd.name;
		elem.querySelector(".memo").textContent = cmd.memo;

		let progs_elem = elem.querySelector(".programs");
		let prog_template = progs_elem.querySelector(":scope > template");

		for(let prog of cmd.programs) {
			let prog_elem = prog_template.content.firstElementChild.cloneNode(true);
			progs_elem.appendChild(prog_elem);
			let prog_text = prog.join(" ");
			prog_elem.querySelector(".line-content").textContent = prog_text;
		}

		elem.addEventListener("click", async (ev) => {

			let status_elems = list.querySelectorAll(".status");
			for(let e of status_elems) {
				e.classList.remove("status-done");
				e.classList.remove("status-inflight");
				e.classList.remove("status-error");
			}

			let target = ev.currentTarget;
			let status_elem = target.querySelector(".status");
			let cmd_name = target.dataset.name;

			status_elem.classList.add("status-inflight");
			let ok = await fetch_ok("/request", { command: cmd_name })
				.then(() => true, () => false);
			status_elem.classList.remove("status-inflight");
			status_elem.classList.add(ok ? "status-done" : "status-error");
		});
	}
});
