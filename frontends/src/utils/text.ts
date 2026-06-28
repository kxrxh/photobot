/**
 * Truncates text to a specified maximum length and adds ellipsis if truncated
 * @param text - The text to truncate
 * @param maxLength - The maximum length before truncation (default: 20)
 * @param suffix - The suffix to add when truncated (default: "...")
 * @returns The truncated text with suffix if needed
 */
export function truncateText(text: string, maxLength = 20, suffix = "..."): string {
	if (!text) return ""
	if (text.length <= maxLength) return text
	return text.substring(0, maxLength - suffix.length) + suffix
}

/**
 * Truncates text for display in UI components with predefined lengths
 */
export const truncate = {
	/** Short truncation for buttons, badges (15 chars) */
	short: (text: string) => truncateText(text, 15),
	/** Medium truncation for cards, lists (30 chars) */
	medium: (text: string) => truncateText(text, 30),
	/** Long truncation for descriptions (50 chars) */
	long: (text: string) => truncateText(text, 50),
	/** Custom truncation with specified length */
	custom: (text: string, length: number) => truncateText(text, length),
}
