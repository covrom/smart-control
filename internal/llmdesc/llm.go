package llmdesc

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/covrom/smart-control/internal/smartdata"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type LLMSmartDescriber struct {
	mu    sync.Mutex
	llm   openai.Client
	model string
}

func NewLLMDescriber(baseUrl, apiKey, model string) *LLMSmartDescriber {
	llm := openai.NewClient(
		option.WithBaseURL(baseUrl), // http://192.168.169.22:37777/v1
		option.WithAPIKey(apiKey),   // 1234
	)
	ret := &LLMSmartDescriber{
		llm:   llm,
		model: model, // qwen3-30b-a3b-instruct-2507 | qvikhr-3-4b-instruction | qwen3-4b-instruct-2507
	}
	return ret
}

const (
	NewDiskSysPrompt = `You are an expert in evaluating hard drive health using S.M.A.R.T. data.  
Carefully analyze the following key parameters from the provided 'smartctl -a' output:  
- Drive temperature  
- Power-on hours (age)  
- Total host reads/writes (data volume handled)  
- Reallocated sectors (Reallocated_Sector_Ct, Reallocated_Event_Count)  
- Uncorrectable errors (UDMA_CRC_Error_Count, Uncorrectable_Error_Count, etc.)  
- Wear leveling count, available spare, percentage used, or other endurance indicators (especially for SSDs/NVMe)  
- Any reallocated, pending, or offline uncorrectable sectors  
- Other critical or warning-level SMART attributes  

Respond **in Russian**, using a concise paragraph followed by a structured summary with emojis:  
✅ = good condition, ⚠️ = requires attention, 🔴 = critical issue  

Use this exact response format:

[Brand] [Model], [Type: HDD/SSD/NVMe], [Form factor], [Memory type if applicable], TBW: [value], MTBF: [value], Speed: [value], Capacity: [value].

✅/⚠️/🔴 [Brief overall assessment highlighting key findings]

🔧 [Actionable advice: e.g., backup data, replace soon, monitor, etc.]

📌 [Critical note if needed: e.g., imminent failure risk, high reallocated sectors, etc.]
`

	CompareSysPrompt = `You are an expert in evaluating storage drive health using S.M.A.R.T. data.
Analyze **both the current and the previous 'smartctl -a' outputs** to assess not only the absolute values but also **trends and changes over time**.

**Important**: A detailed assessment of drive condition is required **only if problems are detected**.
- An increase in recorded data (e.g., power-on hours, total host reads/writes) is **not** a problem by itself — it is normal operational growth.
- Do **not** generate recommendations or detailed analysis if no issues are present.

**Specifically evaluate for signs of degradation or failure**:
- Changes in temperature patterns (unexpected spikes or sustained high temps)
- **Increase** in reallocated sectors, pending sectors, uncorrectable errors, or offline uncorrectable counts
- Degradation in wear leveling count, available spare percentage, or “Percentage Used” (for SSDs/NVMe)
- Any new warning or critical SMART flags that were not present before
- Rate of deterioration (e.g., how fast sectors are being reallocated or spare space is being consumed)

**Response rules**:
- Respond **in Russian**
- If **no problems detected**: Provide only a **brief, one-line assessment** with no recommendations or detailed analysis.
- If **problems detected**: Provide a concise paragraph followed by a structured summary.

Use this exact response format:

[Brand] [Model], [Type: HDD/SSD/NVMe], [Form factor], [Memory type if applicable], TBW: [value], MTBF: [value], Speed: [value], Capacity: [value], Power-on hours: [value], Total written: [terabytes].

✅/⚠️/🔴 [Brief overall assessment — only highlight changes if issues exist.]

🔧 [Actionable advice — **only if problems are present**. Growth in power-on hours or total LBA written (TBW) is expected over time and must not be interpreted as a problem]

📌 [Critical observation about degradation speed, sudden error spikes, or imminent failure risk — **only if problems are present**. Growth in power-on hours or total LBA written (TBW) is expected over time and must not be interpreted as a problem.]
`
)

func (s *LLMSmartDescriber) Describe(ctx context.Context, hostname string, dev, prev smartdata.SMARTDevice) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("llm description begin", "hostname", hostname, "device", dev.Device)

	text := dev.SMARTData

	var messages []openai.ChatCompletionMessageParamUnion

	if prev.SMARTData == "" {
		messages = []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(NewDiskSysPrompt),
			openai.UserMessage(
				fmt.Sprintf("Now analyze the following 'smartctl -a' output and provide your assessment (in Russian):\n\n%s", text),
			),
		}
	} else {
		messages = []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(CompareSysPrompt),
			openai.UserMessage(
				fmt.Sprintf(`Now analyze the following data and provide your assessment (in Russian):

**Previous 'smartctl -a' output:**  
%s

**Current 'smartctl -a' output:**  
%s`,
					prev.SMARTData, text),
			),
		}
	}

	chatCompletion, err := s.llm.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    s.model,
		})
	if err != nil {
		slog.Error("llm.Chat.Completions.New error", "err", err)
		return ""
	}

	slog.Info("llm description done", "hostname", hostname, "device", dev.Device)

	return chatCompletion.Choices[0].Message.Content
}
