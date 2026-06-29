import { useEffect, useState } from "react"
import { TbBulb, TbBulbOff } from "react-icons/tb"

function ThemeSwitch() {
	const [theme, setTheme] = useState(() => {
		const storedTheme = localStorage.getItem("theme")
		if (storedTheme) {
			return storedTheme
		}
		return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
	})

	useEffect(() => {
		const html = document.documentElement
		html.setAttribute("data-theme", theme)
		if (theme === "dark") {
			html.classList.add("dark")
		} else {
			html.classList.remove("dark")
		}
		localStorage.setItem("theme", theme)
	}, [theme])

	const handleToggle = () => {
		setTheme((prevTheme) => (prevTheme === "dark" ? "light" : "dark"))
	}

	return (
		<label className="swap swap-rotate ml-2" aria-label="Переключить тему">
			<input
				type="checkbox"
				className="theme-controller"
				value={theme}
				checked={theme === "dark"}
				onChange={handleToggle}
				aria-label="Тёмная или светлая тема"
			/>
			<TbBulb className="swap-off h-5 w-5" />
			<TbBulbOff className="swap-on h-5 w-5" />
		</label>
	)
}

export default ThemeSwitch
