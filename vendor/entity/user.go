package entity

import (
	"auth"
	"convention/agendaerror"
	"convention/codec"
	"time"
	log "util/logger"
)

// var logln = util.Log
// var logf = util.Logf

type Auth auth.Auth

// Username represents username, a unique identifier, of User
// Identifier
type Username string

// Empty checks if Username empty
func (name Username) Empty() bool {
	return name == ""
}

// Valid checks if Username valid
func (name Username) Valid() bool {
	// FIXME: not only !empty
	return !name.Empty()
}

// TODO: Not sure where to place ...
var allUsersRegistered = *(NewUserList())

// RefInAllUsers returns the ref of a Registered User depending on the Username
func (name Username) RefInAllUsers() *User {
	return allUsersRegistered.Ref(name)
}

// GetAllUsersRegistered returns the reference of the UserList of all Registered Users
func GetAllUsersRegistered() *UserList {
	return &allUsersRegistered
}

type UserIdentifierList []Username

type UserInfoPublic struct {
	Name Username

	Mail  string
	Phone string
}

// UserInfo represents the informations of a User
type UserInfo struct {
	UserInfoPublic

	Auth Auth
}

// UserInfoSerializable represents serializable UserInfo
type UserInfoSerializable = UserInfo

// User represents a User, which is the actor of the operations like sponsor/join/cancel a meeting, etc
type User struct {
	UserInfo
}

// NewUser creates a User object with given UserInfo
func NewUser(info UserInfo) *User {
	if info.Name.Empty() {
		// FIXME: more elegant ?
		// TODO: provide a ready-user that allowe to be empty to be loaded info into ?
		log.Printf("An empty UserInfo is passed to new a User. Just return `nil`.")
		return nil
	}
	u := new(User)
	u.UserInfo = info
	return u
}

// LoadUsersAllRegistered concretely loads all Registered Users
func LoadUsersAllRegistered(decoder codec.Decoder) {
	users := &(allUsersRegistered)
	LoadUserList(decoder, users)
}

// SaveUsersAllRegistered concretely saves all Registered Users
func SaveUsersAllRegistered(encoder codec.Encoder) error {
	users := &(allUsersRegistered)
	return users.Save(encoder)
}

// LoadUser load a User into given container(u) from given decoder
func LoadUser(decoder codec.Decoder, u *User) {
	uInfo := new(UserInfo)
	err := decoder.Decode(uInfo)
	if err != nil {
		log.Fatal(err)
	}
	u.UserInfo = *uInfo
}

// LoadedUser returns loaded User from given decoder
func LoadedUser(decoder codec.Decoder) *User {
	u := new(User)
	LoadUser(decoder, u)
	return u
}

// Save saves User with given encoder
func (u *User) Save(encoder codec.Encoder) error {
	return encoder.Encode(u.UserInfo)
}

func (u *User) Registered() bool {
	if u == nil {
		return false
	}
	return GetAllUsersRegistered().Contains(u.Name)
}

func (u *User) involvedMeetings() *MeetingList {
	return GetAllMeetings().Filter(func(m Meeting) bool {
		return m.SponsoredBy(u.Name) || m.ContainsParticipator(u.Name)
	})
}

func (u *User) FreeWhen(start, end time.Time) bool {
	if u == nil {
		return false
	}

	// NOTE: need improve:
	if err := u.involvedMeetings().ForEach(func(m *Meeting) error {
		s1, e1 := m.StartTime, m.EndTime
		s2, e2 := start, end
		if s1.Before(e2) && e1.After(s2) {
			return agendaerror.ErrConflictedTimeInterval
		}
		return nil
	}); err != nil {
		log.Printf(err.Error())
		return false
	}

	return true
}

// QueryAccount queries an account, where User as the actor
func (u *User) QueryAccount() error {
	return agendaerror.ErrNeedImplement
}

// QueryAccountAll queries all accounts, where User as the actor
func (u *User) QueryAccountAll() UserInfoPublicList {
	// NOTE: FIXME: whatever, temporarily ignore the problem that the actor of query is Nil
	// Hence, now if so, agenda would crash for `Nil.Name`
	ret := GetAllUsersRegistered().PublicInfos()
	log.Printf("User %v queries all accounts.", u.Name)
	return ret
}

// CancelAccount cancels(deletes) the User's own account
func (u *User) CancelAccount() error {
	if u == nil {
		return agendaerror.ErrNilUser
	}
	log.Printf("User %v canceled account.", u.Name)
	return nil
}

// SponsorMeeting creates a meeting, where User as the actor
func (u *User) SponsorMeeting(info MeetingInfo) (*Meeting, error) {
	if u == nil {
		return nil, agendaerror.ErrNilUser
	}
	m := NewMeeting(info)
	err := GetAllMeetings().Add(m)
	log.Printf("User %v sponsors meeting %v.", u.Name, info)
	return m, err
}

// AddParticipatorToMeeting just as its name
func (u *User) AddParticipatorToMeeting(meeting *Meeting, user *User) error {
	if u == nil {
		return agendaerror.ErrNilUser
	}

	err := meeting.Involve(user)
	log.Printf("User %v adds participator %v into Meeting %v.", u.Name, user.Name, meeting.Title)
	return err
}

// RemoveParticipatorFromMeeting just as its name
func (u *User) RemoveParticipatorFromMeeting(meeting *Meeting, user *User) error {
	if u == nil {
		return agendaerror.ErrNilUser
	}
	err := meeting.Exclude(user)
	log.Printf("User %v removes participator %v from Meeting %v.", u.Name, user.Name, meeting.Title)
	return err
}

// LogOut log out User's own (current working) account
func (u *User) LogOut() error {
	if u == nil {
		return agendaerror.ErrNilUser
	}
	return agendaerror.ErrNeedImplement
}

func (u *User) QueryMeetingByInterval(start, end time.Time) MeetingInfoListPrintable {
	// NOTE: FIXME: whatever, temporarily ignore the problem that the actor of query is Nil
	// Hence, now if so, agenda would crash for `Nil.Name`
	ret := u.involvedMeetings().Textualize()
	log.Printf("User %v queries meetings in time interval %v ~ %v.", u.Name, start, end)
	return ret
}

func (u *User) meetingsSponsored() ([]*Meeting, error) {

	return nil, agendaerror.ErrNeedImplement
}

// CancelMeeting cancels(deletes) the given meeting which sponsored by User self, where User as the actor
func (u *User) CancelMeeting(title MeetingTitle) error {
	if u == nil {
		return agendaerror.ErrNilUser
	}
	meeting := title.RefInAllMeetings()
	if meeting == nil {
		return agendaerror.ErrMeetingNotFound
	}

	if !meeting.SponsoredBy(u.Name) {
		return agendaerror.ErrSponsorAuthority
	}

	return meeting.Dissolve()
}

// QuitMeeting quits from the given meeting, where User as the actor
// CHECK: what to do in case User is the sponsor ?
func (u *User) QuitMeeting(title MeetingTitle) error {
	if u == nil {
		return agendaerror.ErrNilUser
	}
	meeting := title.RefInAllMeetings()
	if meeting == nil {
		return agendaerror.ErrMeetingNotFound
	}

	if meeting.SponsoredBy(u.Name) {
		return agendaerror.ErrSponsorResponsibility // NOTE: ???
	}

	if !meeting.ContainsParticipator(u.Name) {
		return agendaerror.ErrUserNotFound
	}

	return meeting.Exclude(u)
}

// ................................................................

// UserList represents a list/group of User (of the form of pointers of Users)
type UserList struct {
	Users map[Username](*User)
}

// UserListRaw also represents a list of User, but it is more trivial and more simple, i.e. it basically is ONLY a list of User, besides this, nothing
// NOTE: these type may be modified/removed in future
type UserListRaw = []*User

type UserInfoPublicList []UserInfoPublic

// UserInfoListSerializable represents a list of serializable UserInfo
type UserInfoListSerializable []UserInfoSerializable

// Serialize just serializes from UserList to UserInfoListSerializable
func (ul *UserList) Serialize() UserInfoListSerializable {
	users := ul.Slice()
	ret := make(UserInfoListSerializable, 0, ul.Size())

	// logln("ul.Size(): ", ul.Size())
	// logf("Serialize: %+v \n", users)
	for _, u := range users {

		// FIXME: these are introduced since up to now, it is possible that UserList contains nil User
		if u == nil {
			log.Printf("A nil User is to be used. Just SKIP OVER it.")
			continue
		}

		ret = append(ret, u.UserInfo)
	}
	return ret
}

// Size just returns the size
func (ulSerial UserInfoListSerializable) Size() int {
	return len(ulSerial)
}

// Deserialize deserializes from serialized UserInfoList to UserList
func (ulSerial UserInfoListSerializable) Deserialize() *UserList {
	ret := NewUserList()

	for _, uInfo := range ulSerial {

		// FIXME: these are introduced since up to now, it is possible that UserList contains nil User
		// FIXME: Not use `== nil` because `uInfo` is a  struct
		if uInfo.Name.Empty() {
			log.Printf("A No-Name UserInfo is to be used. Just SKIP OVER it.")
			continue
		}

		u := NewUser(uInfo)
		if err := ret.Add(u); err != nil {
			log.Printf(err.Error()) // CHECK:
		}
	}
	return ret
}

// NewUserList creates a UserList object
func NewUserList() *UserList {
	ul := new(UserList)
	ul.Users = make(map[Username](*User))
	return ul
}

// NOTE: these API (about loading) may be modified in future
// CHECK: Need in-place load method ?

// LoadUserList loads a UserList into given container(ul) from given decoder
func LoadUserList(decoder codec.Decoder, ul *UserList) {
	// CHECK: Need clear ul ?
	// for decoder.More() {
	// 	user := LoadedUser(decoder)
	// 	if err := ul.Add(user); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	ulSerial := new(UserInfoListSerializable)
	if err := decoder.Decode(ulSerial); err != nil {
		log.Fatal(err)
	}
	for _, uInfoSerial := range *ulSerial {
		u := NewUser(uInfoSerial)
		if err := ul.Add(u); err != nil {
			log.Printf(err.Error())
		}
	}
}

// InitFrom loads UserList in-place from given UserIdentifierList; Just like `init`
func (ul *UserList) InitFrom(li UserIdentifierList) {
	// clear ...
	ul.Users = NewUserList().Users

	for _, id := range li {
		u := id.RefInAllUsers() // CHECK: ditto
		ul.Add(u)
	}
}

// LoadFrom loads UserList in-place from given decoder; Just like `init`
func (ul *UserList) LoadFrom(decoder codec.Decoder) {
	LoadUserList(decoder, ul)
}

// LoadedUserList returns loaded UserList from given decoder
func LoadedUserList(decoder codec.Decoder) *UserList {
	ul := NewUserList()
	LoadUserList(decoder, ul)
	return ul
}

func (ul *UserList) identifiers() []Username {
	ret := make([]Username, 0, ul.Size())
	for _, u := range ul.PublicInfos() {
		ret = append(ret, u.Name)
	}
	return ret
}
func (ul *UserList) PublicInfos() UserInfoPublicList {
	users := ul.Slice()
	ret := make(UserInfoPublicList, 0, ul.Size())

	for _, u := range users {

		// FIXME: these are introduced since up to now, it is possible that UserList contains nil User
		if u == nil {
			log.Printf("A nil User is to be used. Just SKIP OVER it.")
			continue
		}

		ret = append(ret, u.UserInfoPublic)
	}
	return ret
}

// Save use given encoder to Save UserList
func (ul *UserList) Save(encoder codec.Encoder) error {
	sl := ul.Serialize()
	// logf("sl: %+v\n", sl)
	return encoder.Encode(sl)
}

// Size just returns the number of User reference in UserList
func (ul *UserList) Size() int {
	return len(ul.Users)
}

// Ref just returns the reference of user with the given name
func (ul *UserList) Ref(name Username) *User {
	return ul.Users[name] // NOTE: if directly return accessed result from a map like this, would not get the (automatical) `ok`
}

// Contains just check if contains
func (ul *UserList) Contains(name Username) bool {
	u := ul.Ref(name)
	return u != nil
}

// Add just add
func (ul *UserList) Add(user *User) error {
	if user == nil {
		return agendaerror.ErrNilUser
	}
	name := user.Name
	if ul.Contains(name) {
		return agendaerror.ErrExistedUser
	}
	ul.Users[name] = user
	return nil
}

// Remove just remove
func (ul *UserList) Remove(user *User) error {
	if user == nil {
		return agendaerror.ErrNilUser
	}
	name := user.Name
	if ul.Contains(name) {
		delete(ul.Users, name) // NOTE: never error, according to 'go-maps-in-action'
		return nil
	}
	return agendaerror.ErrUserNotFound
}

// PickOut =~= Ref and then Remove
func (ul *UserList) PickOut(name Username) (*User, error) {
	if name.Empty() {
		return nil, agendaerror.ErrEmptyUsername
	}
	u := ul.Ref(name)
	if u == nil {
		return u, agendaerror.ErrUserNotFound
	}
	defer ul.Remove(u)
	return u, nil
}

// Slice returns a UserListRaw based on UserList ul
// NOTE: for the need of this simple agenda system, this seems somewhat needless
func (ul *UserList) Slice() UserListRaw {
	users := make(UserListRaw, 0, ul.Size())
	for _, u := range ul.Users {
		users = append(users, u) // CHECK: maybe better to use index in golang ?
	}
	return users
}

// ForEach used for all extension/concrete logic for whole UserList
func (ul *UserList) ForEach(fn func(*User) error) error {
	for _, v := range ul.Users {
		if err := fn(v); err != nil {
			// CHECK: Or, lazy error ?
			return err
		}
	}
	return nil
}

// Filter used for all extension/concrete select for whole UserList
func (ul *UserList) Filter(pred func(User) bool) *UserList {
	ret := NewUserList()
	for _, u := range ul.Users {
		if pred(*u) {
			ret.Add(u)
		}
	}
	return ret
}
