import { cleanup, render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { getImageDataUrl } from "@/utils/image"
import { ObjectImage } from "./ObjectImage"

vi.mock("@/utils/image", () => ({
	getImageDataUrl: vi.fn(),
}))

const defaultObject = {
	id: "1",
	presigned_url: "https://example.com/image.jpg",
}

const defaultProps = {
	object: defaultObject,
	isSelected: false,
	isControlModeActive: true,
	onClick: vi.fn(),
}

describe("ObjectImage", () => {
	beforeEach(() => {
		vi.mocked(getImageDataUrl).mockReset()
		vi.mocked(getImageDataUrl).mockReturnValue("data:image/jpeg;base64,/9j/4AAQ")
	})

	afterEach(() => {
		cleanup()
	})

	it("renders image when getImageDataUrl returns URL", () => {
		render(<ObjectImage {...defaultProps} />)
		expect(screen.getByRole("img", { name: "1" })).toBeInTheDocument()
	})

	it("renders placeholder icon when getImageDataUrl returns null", () => {
		vi.mocked(getImageDataUrl).mockReturnValue(null)
		render(<ObjectImage {...defaultProps} />)
		expect(screen.queryByRole("img")).not.toBeInTheDocument()
		// No semantic role on the vector placeholder; scope under the tile button.
		expect(screen.getByRole("button").querySelector("svg")).toBeInTheDocument()
	})

	it("shows checkmark overlay when selected in select mode", () => {
		render(<ObjectImage {...defaultProps} isSelected mode="select" />)
		const button = screen.getByRole("button")
		expect(button.querySelector("img[alt='1']")).toBeInTheDocument()
		expect(button.querySelector(".bg-primary\\/50")).toBeInTheDocument()
	})

	it("shows X overlay when excluded in exclude mode", () => {
		render(<ObjectImage {...defaultProps} isExcluded mode="exclude" />)
		expect(screen.getByRole("button").querySelector(".bg-base-content\\/45")).toBeInTheDocument()
	})

	it("calls onClick when clicked", async () => {
		const onClick = vi.fn()
		render(<ObjectImage {...defaultProps} onClick={onClick} />)
		await userEvent.click(screen.getByRole("button"))
		expect(onClick).toHaveBeenCalledTimes(1)
	})

	it("uses presigned_url over file when both present", () => {
		const obj = { id: "2", presigned_url: "https://presigned.jpg", file: "legacy" }
		render(<ObjectImage {...defaultProps} object={obj} />)
		expect(getImageDataUrl).toHaveBeenCalledWith("https://presigned.jpg")
	})
})
