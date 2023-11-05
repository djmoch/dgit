function doHash() {
	if (event.oldURL.split("#").length === 2) {
		oldId = event.oldURL.split("#")[1];
		document.getElementById(oldId).bgColor = "#ffd";
	}
	newId = event.newURL.split("#")[1];
	document.getElementById(newId).bgColor = "#eea";
}

function initHash(hash) {
	if (window.location.hash != "") {
		id = hash.split("#")[1];
		document.getElementById(id).bgColor = "#eea";
	}
}

window.addEventListener("hashchange", doHash);
