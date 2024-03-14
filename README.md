## Introduction

This is an educational LINE Bot that help badminton teachers keep track of their students' learning progress. It makes use of LINE's rich menu and postback action to provide a more interactive user experience. Besides, with firebase, Google Drive API, LINE Messaging API, and an internal API to an pose estimation AI model, it connects all the necessary services, qualifying itself to be an agent of teachers. This project has 2 branches:

- `main`: the project with more functionalities and features
- `moe`: the project with less functionalities and features, but with a more concise codebase

## Postback Action Data

### parameters

- `type`
  - analyze_video
  - add_reflection
  - view_portfolio
  - add_preview_note
  - view_instruction
  - view_expert_video
- `handedness`
  - left
  - right
- `date`
  - YYYY-MM-DD-HH-MM
- `skill`
  - serve
  - smash
  - clear
- `video_id`
