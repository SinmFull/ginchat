package service

import (
	"fmt"
	"ginchat/models"
	"ginchat/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	// "github.com/asaskevich/govalidator"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetUserList
// Summary  所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data = models.GetUserList()
	c.JSON(200, gin.H{
		"code":    0, //0正确 -1 错误
		"message": "获取用户列表成功",
		"data":    data,
	})
}

// FindUserByNameAndPwd
// Summary  所有用户
// @Tags 用户模块
// @param	name query string false "用户名"
// @param	password query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}

	// name := c.Query("name")
	// password := c.Query("password")
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, //0正确 -1 错误
			"message": "用户不存在",
			"data":    data,
		})
		return
	}

	flag := utils.ValidPassword(password, user.Salt, user.PassWord)
	if !flag {
		c.JSON(200, gin.H{
			"code":    -1, //0正确 -1 错误
			"message": "密码错误",
			"data":    data,
		})
		return
	}
	pwd := utils.MakePassword(password, user.Salt)
	data = models.FindUserByNameAndPwd(name, pwd)
	c.JSON(200, gin.H{
		"code":    0, //0正确 -1 错误
		"message": "登录成功",
		"data":    data,
	})
}

// CreateUser
// Summary 新增用户
// @Tags 用户模块
// @param	name query string false "用户名"
// @param	password query string false "密码"
// @param	repassword query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	Identity := c.Request.FormValue("Identity")

	salt := fmt.Sprintf("%06d", rand.Int31())

	data := models.FindUserByName(user.Name)

	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, //0正确 -1 错误
			"message": "用户名不能为空",
			"data":    data,
		})
		return
	}
	if password == "" {
		c.JSON(200, gin.H{
			"code":    -1, //0正确 -1 错误
			"message": "密码不能为空",
			"data":    data,
		})
		return
	}
	if data.Name != "" {
		c.JSON(200, gin.H{
			"code":    -1, //0正确 -1 错误
			"message": "用户名已注册",
			"data":    data,
		})
		return
	}
	if password != Identity {
		c.JSON(200, gin.H{
			"code":    -1, //0正确 -1 错误
			"message": "两次密码不一致",
			"data":    data,
		})
		return
	}
	//user.PassWord = password
	user.PassWord = utils.MakePassword(password, salt)
	user.Salt = salt
	models.CreateUser(user)
	c.JSON(200, gin.H{
		"code":    0, //0正确 -1 错误
		"message": "新增用户成功",
		"data":    data,
	})
}

// DeleteUser
// Summary 删除用户
// @Tags 用户模块
// @param	id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    0, //0正确 -1 错误
		"message": "删除用户成功",
		"data":    user,
	})
}

// DeleteUser
// Summary 更新用户
// @Tags 用户模块
// @param	id formData string false "id"
// @param	name formData string false "name"
// @param	password formData string false "password"
// @param	phone formData string false "phone"
// @param	email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.PassWord = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1, //0正确 -1 错误
			"message": "修改参数不匹配",
			"data":    user,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(200, gin.H{
			"code":    0, //0正确 -1 错误
			"message": "修改用户成功",
			"data":    user,
		})
	}

}

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("err")
	}

	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)

	MsgHandler(ws, c)
}

func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.PostForm("userIdA"))
	userIdB, _ := strconv.Atoi(c.PostForm("userIdB"))
	start, _ := strconv.Atoi(c.PostForm("start"))
	end, _ := strconv.Atoi(c.PostForm("end"))
	isRev, _ := strconv.ParseBool(c.PostForm("isRev"))
	res := models.RedisMsg(int64(userIdA), int64(userIdB), int64(start), int64(end), isRev)
	utils.RespOKList(c.Writer, "ok", res)
}

func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	for {
		msg, err := utils.Subscribe(c, utils.Publishkey)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("发送消息：", msg)
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func SearchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SearchFriend(uint(id))
	// c.JSON(200, gin.H{
	// 	"code":    0, //0正确 -1 错误
	// 	"message": "查询好友列表成功",
	// 	"data":    users,
	// })
	utils.RespOKList(c.Writer, users, len(users))
}

func AddFriend(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Request.FormValue("userId"))
	tname := c.Request.FormValue("targetName")

	fmt.Println(tname)
	code, msg := models.AddFriend(uint(uid), tname)
	// c.JSON(200, gin.H{
	// 	"code":    0, //0正确 -1 错误
	// 	"message": "查询好友列表成功",
	// 	"data":    users,
	// })
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}

}

func CreateCommunity(c *gin.Context) {
	oid, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	name := c.Request.FormValue("name")
	community := models.Community{}
	community.OwnerId = uint(oid)
	community.Name = name

	code, msg := models.CreateCommunity(community)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func LoadCommunity(c *gin.Context) {
	oid, _ := strconv.Atoi(c.Request.FormValue("ownerId"))

	data, msg := models.LoadCommunity(uint(oid))
	if len(data) != 0 {
		utils.RespList(c.Writer, 0, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func JoinGroup(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Request.FormValue("userId"))
	comid := c.Request.FormValue("comId")
	//comId := c.Request.FormValue("targetName")
	code, msg := models.JoinGroup(uint(uid), comid)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}

}
