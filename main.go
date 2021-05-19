package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	Token string
	Lobby map[string]LobbyDetails
	RaidInfo map[string]string
)

type LobbyDetails struct {
	RoleName string
	RoleID string
	LeaderID string
	Raiders []string
	//Location string
	Time     string
	MaxInvites int
	Boss      string
	Ready     string
}

func init() {
	flag.StringVar(&Token, "t", os.Getenv("DISCORD_TOKEN"), "Bot Token")
	flag.Parse()

	Lobby = make(map[string]LobbyDetails)
	RaidInfo = make(map[string]string)
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)

	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(emojiReact)
	dg.AddHandler(emojiRemove)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func emojiRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {

	/*messageID := m.MessageID
	lobDetails := Lobby[messageID]

	fmt.Println(m.UserID)
	fmt.Println(m.MessageReaction.UserID)

	fmt.Println(lobDetails.Raiders)

	ind := indexOf(lobDetails.Raiders, m.UserID)

	fmt.Println(ind)
	raiders := lobDetails.Raiders

	raiders[ind] = raiders[len(raiders)-1]
	raiders[len(raiders)-1] = ""
	lobDetails.Raiders = raiders[:len(raiders)-1]
	Lobby[messageID] = lobDetails

	guildie, _ := s.GuildMember(m.GuildID, m.UserID)
	roleInd := indexOf(guildie.Roles, lobDetails.RoleID)
	roles := guildie.Roles

	roles[roleInd] = roles[len(roles)-1]
	roles[len(roles)-1] = ""
	roles = roles[:len(roles)-1]

	s.GuildMemberEdit(m.GuildID, m.UserID, roles)

	fmt.Println(lobDetails) */

}
func emojiReact(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.UserID == s.State.User.ID {
		return
	}

	if m.Emoji.Name == "✅" {
		messageID := m.MessageID
		lobDetails := Lobby[messageID]

		if len(lobDetails.Raiders) < lobDetails.MaxInvites {
			if !isInSlice(lobDetails.Raiders, m.UserID) {
				lobDetails.Raiders = append(lobDetails.Raiders, m.UserID)
				Lobby[messageID] = lobDetails

				guildie, _ := s.GuildMember(m.GuildID, m.UserID)
				s.GuildMemberEdit(m.GuildID, m.UserID, append(guildie.Roles, lobDetails.RoleID))
			}

			fmt.Println(lobDetails)
		} else {
			fmt.Println("it's fucking full wtf...")
		}
	}

	if m.Emoji.Name == "❌" {
		messageID := m.MessageID
		lobDetails := Lobby[messageID]

		if m.UserID != lobDetails.LeaderID {
			return
		}

		fmt.Println("Lobby closing...")
		fmt.Println(s.ChannelMessageDelete(m.ChannelID, messageID))
		time.Sleep(1 * time.Second)
		fmt.Println(s.ChannelMessageDelete(m.ChannelID, lobDetails.Ready))
		s.GuildRoleDelete(m.GuildID, lobDetails.RoleID)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	//bob, _ := s.GuildMember(m.GuildID, m.Author.ID)


	if strings.Contains(m.Content, ".set") {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		values := strings.Split(m.Content, " ")
		poke := values[1]
		data := values[2]

		RaidInfo[poke] = data

		msg, _ :=s.ChannelMessageSend(m.ChannelID, poke + "'s raid info has been saved!")
		time.Sleep(3 * time.Second)
		s.ChannelMessageDelete(m.ChannelID, msg.ID)

	}

	if strings.Contains(m.Content, ".ready") {
		values := strings.Split(m.Content, " ")
		role := values[1]
		roleID := ""
		roles, _ := s.GuildRoles(m.GuildID)
		var raidMembers []string
		for _, ele := range roles {
			if ele.Mention() == role {
				fmt.Println(Lobby)
				roleID = ele.ID
				members, _ := s.GuildMembers(m.GuildID, "", 0)
				for _, member := range members {
					if isInSlice(member.Roles, ele.ID) {
						fmt.Println(member.Nick + " is here!!!")
						raidMembers = append(raidMembers, member.Nick)
					}
				}
			}
		}

		printString := ""

		for _, ele := range raidMembers {
			printString += ele + "\n"
		}

		messageEmbed := &discordgo.MessageEmbed{}
		member, _ := s.GuildMember(m.GuildID, m.Author.ID)

		messageEmbed.Author = &discordgo.MessageEmbedAuthor{
			Name:    member.Nick,
			IconURL: m.Author.AvatarURL(""),
		}

		messageEmbed.Title = "READY UP"
		messageEmbed.Description = "The raid lobby is starting!"
		messageEmbed.Color = 50

		messageEmbed.Fields = []*discordgo.MessageEmbedField{
			{
				Name: "Raid Party",
				Value: printString,
			},
		}

		messageEmbed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Built by iG4ymer",
		}

		message := &discordgo.MessageSend {
			Embed: messageEmbed,
		}

		msg, err := s.ChannelMessageSendComplex(m.ChannelID, message)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(msg)

		for k, v := range Lobby {
			if v.RoleID == roleID {
				lob := Lobby[k]
				lob.Ready = msg.ID
				Lobby[k] = lob
			}
		}

		fmt.Println(Lobby)
	}

	//.r Pikachu 5:00 4 <pokemon, time, invites available>
	if strings.Contains(m.Content, ".r") {
		s.ChannelMessageDelete(m.ChannelID, m.ID)

		values := strings.Split(m.Content, " ")
		values = values[1:]

		if len(values) == 3 {
			//location := values[1]
			boss := values[0]
			t := values[1]
			numInvites, _ := strconv.Atoi(values[2])

			layout := "15:04"

			startTime, err := time.Parse(layout, t)
			endTime := startTime.Add(time.Minute * 45)

			t = startTime.Format(layout) + " - " + endTime.Format(layout)

			messageEmbed := &discordgo.MessageEmbed{}
			member, _ := s.GuildMember(m.GuildID, m.Author.ID)

			messageEmbed.Author = &discordgo.MessageEmbedAuthor{
				Name:    member.Nick,
				IconURL: m.Author.AvatarURL(""),
			}

			messageEmbed.Title = boss

			if _, ok := RaidInfo[boss]; ok {
				messageEmbed.URL = RaidInfo[boss]
			}

			messageEmbed.Description = member.Nick + " has " + values[2] + " invites for this raid. Please ✅ to sign up. First come, first serve. \n" +
				"When raid is complete, please ❌. Thanks!"
			messageEmbed.Color = 50

			role := member.Nick + "-" + boss + "-" + getNewID()

			messageEmbed.Fields = []*discordgo.MessageEmbedField{
				{
					Name: "Time",
					Value: t,
				},
				/*{
					Name: "Location",
					Value: location,
				}, */
				{
					Name: "Role to mention before entering raid:",
					Value: "@" + role,
				},
			}

			messageEmbed.Footer = &discordgo.MessageEmbedFooter{
				Text: "Built by iG4ymer",
			}

			message := &discordgo.MessageSend {
				Embed: messageEmbed,
			}

			sentMessage, err := s.ChannelMessageSendComplex(m.ChannelID, message)

			if err != nil {
				fmt.Println(err)
			}

			s.MessageReactionAdd(m.ChannelID, sentMessage.ID, "✅")
			s.MessageReactionAdd(m.ChannelID, sentMessage.ID, "❌")

			r, _ := s.GuildRoleCreate(m.GuildID)
			s.GuildRoleEdit(m.GuildID, r.ID, role, 0, false, 0, true)


			lob := LobbyDetails{
				role,
				r.ID,
				m.Author.ID,
				[]string{},
				//location,
				t,
				numInvites,
				boss,
				"",
			}

			Lobby[sentMessage.ID] = lob
		}
	}
}



//Utilities
func getNewID() string {

	f := make([]byte, 2)
	_, err := rand.Read(f)

	if err != nil {
		log.Fatal(err)
	}

	uuid := fmt.Sprintf("%x", f[0:2])
	return uuid
}

func isInSlice(slice []string, item string) bool {
	for _, ele := range slice {
		if strings.EqualFold(ele, item) {
			return true
		}
	}

	return false
}

func indexOf(slice []string, item string) int {
	for ind, ele := range slice {
		fmt.Println(ele, item)
		if strings.EqualFold(ele, item) {
			return ind
		}
	}

	return -1
}
