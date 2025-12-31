package service

import (
	"GinChat/models"
	"GinChat/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

// GetUserList
// @Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data = models.GetUserList()
	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "获取用户列表成功！",
		"data":    data,
	})
}

// CreateUser
// @Summary 新增用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @param rePassword query string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	rePassword := c.Request.FormValue("Identity")
	salt := fmt.Sprintf("%06d", rand.Int31())
	if user.Name == "" || password == "" || rePassword == "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "用户名或密码不能为空！",
			"data":    user,
		})
		return
	}
	if password != rePassword {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "两次密码不一致！",
			"data":    user,
		})
		return
	}

	// check duplicate name
	_, err := models.FindUserByName(user.Name)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "用户名已注册！",
			"data":    user,
		})
		return
	}

	user.PassWord = utils.MakePassword(password, salt)
	user.Salt = salt
	user.LoginTime = time.Now()
	user.LoginOutTime = time.Now()
	user.HeartbeatTime = time.Now()
	models.CreateUser(user)
	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "新增用户成功！",
		"data":    user,
	})
}

// FindUserByNameAndPwd
// @Summary 用户登录
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	user, err := models.FindUserByName(name)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "该用户不存在",
			"data":    nil,
		})
		return
	}

	if !utils.ValidPassword(password, user.Salt, user.PassWord) {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "密码不正确",
			"data":    nil,
		})
		return
	}
	user.Identity, err = utils.GenerateTokens(user.ID, user.Name)
	if err != nil {
		fmt.Println("generate user token error: ", err)
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "error generate fail",
			"data":    nil,
		})
		return
	}
	// (choose)
	err = models.RefreshSQLToken(user, user.Identity)
	if err != nil {
		fmt.Println("refresh database data error: ", err)
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "error refresh database data",
			"data":    user,
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "登录成功",
		"data":    user,
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户模块
// @param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "删除用户成功！",
		"data":    user,
	})

}

// UpdateUser
// @Summary 修改用户
// @Tags 用户模块
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.PassWord = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Avatar = c.PostForm("icon")
	user.Email = c.PostForm("email")

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "邮箱/电话号码格式错误！",
			"data":    user,
		})
		return
	}

	origin := models.FindByID(user.ID)

	// check duplicate (name, email, phone)
	_, err = models.FindUserByName(user.Name)
	if err == nil && origin.ID != user.ID { // 查到记录且不是自己 → 重复
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "用户名已被使用",
			"data":    user,
		})
		return
	}
	_, err = models.FindUserByEmail(user.Email)
	if err == nil && origin.Email != user.Email {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "邮箱已绑定",
			"data":    user,
		})
		return
	}
	_, err = models.FindUserByPhone(user.Phone)
	if err == nil && origin.Phone != user.Phone {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "手机号码已绑定",
			"data":    user,
		})
		return
	}

	// check success
	user.PassWord = utils.MakePassword(user.PassWord, origin.Salt)
	result := models.UpdateUser(user)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "服务器出错",
			"data":    user,
		})
		panic(fmt.Sprintf("UpdateUser -- 修改数据库错误：%v", result.Error))
		return
	}
	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "修改用户成功！",
		"data":    user,
	})
}

// 防止跨域站点伪造请求
//var upGrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//}
//
//func SendMsg(c *gin.Context) {
//	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	defer func(ws *websocket.Conn) {
//		err = ws.Close()
//		if err != nil {
//			fmt.Println(err)
//		}
//	}(ws)
//	MsgHandler(c, ws)
//}
//
//func MsgHandler(c *gin.Context, ws *websocket.Conn) {
//	for {
//		msg, err := utils.Subscribe(c, utils.PublishKey)
//		if err != nil {
//			fmt.Println(" MsgHandler 发送失败", err)
//		}
//
//		tm := time.Now().Format("2006-01-02 15:04:05")
//		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
//		err = ws.WriteMessage(1, []byte(m))
//		if err != nil {
//			log.Fatalln(err)
//		}
//	}
//}

func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.PostForm("userIdA"))
	userIdB, _ := strconv.Atoi(c.PostForm("userIdB"))
	start, _ := strconv.Atoi(c.PostForm("start"))
	end, _ := strconv.Atoi(c.PostForm("end"))
	isRev, _ := strconv.ParseBool(c.PostForm("isRev"))
	res := models.RedisMsg(int64(userIdA), int64(userIdB), int64(start), int64(end), isRev)
	utils.RespOKList(c.Writer, "ok", res)
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

// SearchFriends 查找好友列表
func SearchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SearchFriend(uint(id))
	utils.RespOKList(c.Writer, users, len(users))
}

// AddFriend 添加好友
func AddFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetName := c.Request.FormValue("targetName")
	code, msg := models.AddFriend(uint(userId), targetName)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// CreateCommunity 新建群
func CreateCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	community := models.Community{
		OwnerId: uint(ownerId),
		Name:    c.Request.FormValue("name"),
		Img:     c.Request.FormValue("img"),
		Desc:    c.Request.FormValue("desc"),
	}
	code, msg := models.CreateCommunity(community)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// LoadCommunity 加载群列表
func LoadCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	//	name := c.Request.FormValue("name")
	data, msg := models.LoadCommunity(uint(ownerId))
	if len(data) != 0 {
		utils.RespList(c.Writer, 0, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// JoinGroups 加入群 userId uint, comId uint
func JoinGroups(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	comId := c.Request.FormValue("comId")

	//	name := c.Request.FormValue("name")
	data, msg := models.JoinGroup(uint(userId), comId)
	if data == 0 {
		utils.RespOK(c.Writer, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// FindByID 根据用户Id查找用户
func FindByID(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	data := models.FindByID(uint(userId))
	utils.RespOK(c.Writer, data, "ok")
}
