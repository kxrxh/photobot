import { Button } from "@/components/common/ui/Button"
import { MESSENGER_RELOAD_MESSAGE } from "@/lib/auth/messages"

type MessengerReloadPromptProps = {
	message?: string | null
}

export function MessengerReloadPrompt({ message }: MessengerReloadPromptProps) {
	return (
		<div className="flex flex-col items-center justify-center gap-4 p-4 min-h-screen">
			<p className="text-center text-base-content/80">{message ?? MESSENGER_RELOAD_MESSAGE}</p>
			<Button type="button" variant="primary" onClick={() => window.location.reload()}>
				Перезапустить
			</Button>
		</div>
	)
}
