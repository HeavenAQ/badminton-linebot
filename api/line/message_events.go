package line

import (
	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) SendDefaultReply(replyToken string) {
	handler.bot.ReplyMessage(replyToken, &linebot.TextMessage{Text: "請點選選單的項目"})
}

func (handler *LineBotHandler) SendInstruction(replyToken string) {
	const welcome = "歡迎加入羽球教室🏸，以下為選單的使用說明:\n\n"
	const instruction = "➡️ 使用說明：呼叫選單各個項目的解說\n\n"
	const portfolio = "➡️ 我的學習歷程：查看個人每周的學習歷程記錄\n\n"
	const expertVideo = "➡️ 專家影片：觀看專家示範影片\n\n"
	const uploadRecording = "➡️ 上傳錄影：上傳個人動作錄影\n\n"
	const addPortfolio = "➡️ 新增學習反思：新增每周各動作的學習反思\n\n"
	const syllabus = "➡️ 課程大綱：查看課程大綱\n\n"
	const note = "⚠️ 每周的學習歷程都需要當週的【影片】以及【學習反思】才能建檔"
	const msg = welcome + instruction + portfolio + expertVideo + uploadRecording + addPortfolio + syllabus + note
	handler.bot.ReplyMessage(replyToken, &linebot.TextMessage{Text: msg})
}

func (handler *LineBotHandler) SendSyllabus(replyToken string) {
	const syllabus = "課程大綱\n\n"
	const msg = syllabus + "https://test.com"
	handler.bot.ReplyMessage(replyToken, &linebot.TextMessage{Text: msg})
}

func (handler *LineBotHandler) PromptSelectReflection(replyToken string, user db.UserData) {
	msg := linebot.NewTextMessage("請選擇要新增學習反思的動作").WithQuickReplies(
		handler.getSkillQuickReplyItems(AddReflection),
	)
	handler.bot.ReplyMessage(replyToken, msg)
}
