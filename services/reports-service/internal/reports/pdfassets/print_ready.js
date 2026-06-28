// Waits for fonts, images, and layout paint before PrintToPDF.
(async () => {
	const maxWaitMs = 60000;
	const deadline = new Promise((_, reject) =>
		setTimeout(
			() => reject(new Error(`report print ready: timeout after ${maxWaitMs}ms`)),
			maxWaitMs,
		),
	);
	const waitImage = (img) => {
		if (img.complete) return Promise.resolve();
		return new Promise((resolve) => {
			const done = () => resolve();
			img.addEventListener('load', done, { once: true });
			img.addEventListener('error', done, { once: true });
		});
	};
	const ready = (async () => {
		if (document.fonts?.ready) {
			await document.fonts.ready;
		}
		await Promise.all(Array.from(document.images || []).map(waitImage));
		await new Promise((resolve) =>
			requestAnimationFrame(() => requestAnimationFrame(resolve)),
		);
	})();
	await Promise.race([deadline, ready]);
})();
