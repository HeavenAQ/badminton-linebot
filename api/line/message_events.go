package line

import (
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) SendReply(replyToken string, msg string) (*linebot.BasicResponse, error) {
	return handler.bot.ReplyMessage(replyToken, linebot.NewTextMessage(msg)).Do()
}

func (handler *LineBotHandler) SendDefaultReply(replyToken string) (*linebot.BasicResponse, error) {
	return handler.SendReply(replyToken, "請點選選單的項目")
}

func (handler *LineBotHandler) SendDefaultErrorReply(replyToken string) (*linebot.BasicResponse, error) {
	return handler.SendReply(replyToken, "發生錯誤，請重新操作")
}

func (handler *LineBotHandler) SendWelcomeReply(event *linebot.Event) (*linebot.BasicResponse, error) {
	username, err := handler.GetUserName(event.Source.UserID)
	if err != nil {
		return nil, err
	}
	welcomMsg := "Hi " + username + "! 歡迎加入羽球教室🏸\n" + "已建立您的使用者資料🎉🎊 請點選選單的項目開始使用"
	return handler.SendReply(event.ReplyToken, welcomMsg)
}

func (handler *LineBotHandler) SendVideoUploadedReply(replyToken string, skill string, videoFolder string) (*linebot.BasicResponse, error) {
	s := SkillStrToEnum(skill)
	skillFolder := "https://drive.google.com/drive/u/0/folders/" + videoFolder
	return handler.bot.ReplyMessage(
		replyToken,
		linebot.NewTextMessage("已成功上傳影片!"),
		linebot.NewTextMessage("以下為【"+s.ChnString()+"】的影片資料夾：\n"+skillFolder),
	).Do()
}

func (handler *LineBotHandler) SendInstruction(replyToken string) (*linebot.BasicResponse, error) {
	const welcome = "歡迎加入羽球教室🏸，以下為選單的使用說明:\n\n"
	const instruction = "➡️ 使用說明：呼叫選單各個項目的解說\n\n"
	const portfolio = "➡️ 學習歷程：查看個人每周的學習歷程記錄\n\n"
	const expertVideo = "➡️ 專家影片：觀看專家示範影片\n\n"
	const addPreviewNote = "➡️ 課前動作檢測：課前預習上週動作，並記錄需進步的要點\n\n"
	const analyzeRecording = "➡️ 分析影片：上傳個人動作錄影，系統將自動產生分析結果\n\n"
	const addReflection = "➡️ 本週學習反思：新增每周各動作的學習反思\n\n"
	const note = "⚠️ 每周的學習歷程都需有【影片】才能建檔"
	const msg = welcome + instruction + portfolio + expertVideo + addPreviewNote + analyzeRecording + addReflection + note
	return handler.bot.ReplyMessage(replyToken, linebot.NewTextMessage(msg)).Do()
}

func (handler *LineBotHandler) SendSyllabus(replyToken string) (*linebot.BasicResponse, error) {
	const syllabus = "課程大綱：\n"
	const msg = syllabus + "https://test.com"
	return handler.bot.ReplyMessage(replyToken, linebot.NewTextMessage(msg)).Do()
}

func (handler *LineBotHandler) PromptSkillSelection(replyToken string, action Action, prompt string) (*linebot.BasicResponse, error) {
	msg := linebot.NewTextMessage(prompt).WithQuickReplies(
		handler.getSkillQuickReplyItems(action),
	)
	return handler.bot.ReplyMessage(replyToken, msg).Do()
}

func (handler *LineBotHandler) PromptHandednessSelection(replyToken string) (*linebot.BasicResponse, error) {
	msg := linebot.NewTextMessage("請選擇左手或右手").WithQuickReplies(
		handler.getHandednessQuickReplyItems(),
	)
	return handler.bot.ReplyMessage(replyToken, msg).Do()
}

func (handler *LineBotHandler) SendVideoMessage(replyToken string, video VideoInfo) (*linebot.BasicResponse, error) {
	videoLink := "https://drive.google.com/uc?export=download&id=" + video.VideoId
	thumbnailLink := "https://drive.usercontent.google.com/download?id=" + video.ThumbnailId
	return handler.bot.ReplyMessage(
		replyToken,
		linebot.NewVideoMessage(videoLink, thumbnailLink),
	).Do()
}
