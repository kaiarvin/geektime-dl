package cmds

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mmzou/geektime-dl/cli/application"
	"github.com/mmzou/geektime-dl/downloader"
	"github.com/mmzou/geektime-dl/service"
	"github.com/mmzou/geektime-dl/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

// NewCourseCommand login command
func NewCourseCommand() []cli.Command {
	return []cli.Command{
		{
			Name:      "column",
			Usage:     "获取专栏列表",
			UsageText: appName + " column",
			Action:    columnAction,
		},
		{
			Name:      "video",
			Usage:     "获取视频课程列表",
			UsageText: appName + " video",
			Action:    videoAction,
		},
		{
			Name:   "dlall",
			Usage:  "下载所有课件",
			Action: DownloadAllColumn,
		},
	}
}

func columnAction(c *cli.Context) error {
	columns, err := application.Columns()
	if err != nil {
		return err
	}

	renderCourses(columns)

	return nil
}

func videoAction(c *cli.Context) error {
	videos, err := application.Videos()
	if err != nil {
		return err
	}

	renderCourses(videos)

	return nil
}

func renderCourses(courses []*service.Course) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "ID", "名称", "时间", "作者", "购买"})
	table.SetAutoWrapText(false)

	for i, p := range courses {
		isBuy := ""
		if p.HadSub {
			isBuy = "是"
		}
		table.Append([]string{strconv.Itoa(i), strconv.Itoa(p.ID), p.ColumnTitle, time.Unix(int64(p.ColumnCtime), 0).Format("2006-01-02"), p.AuthorName, isBuy})
	}

	table.Render()
}

func DownloadAllColumn(c *cli.Context) {
	columns, err := application.Columns()
	if err != nil {
		fmt.Println("GetColumn Columns:", err)
		return
	}

	renderCourses(columns)
	var have bool = false
	for index, column := range columns {
		fmt.Println("第 ", index, "/", len(columns), " 名称:", column.ColumnTitle)

		cid := column.ID
		course, articles, err := application.CourseWithArticles(cid)
		if err != nil {
			fmt.Println("GetColumn CourseWithArticles:", err)
			return
		}

		downloadData := extractDownloadData(course, articles, 0)

		if _info {
			downloadData.PrintInfo()
			return
		}

		sub := "MP4"
		if course.IsColumn() {
			sub = "MP3"
		}

		path, err := utils.Mkdir(utils.FileName(course.ColumnTitle, ""), sub)

		errors := make([]error, 0)
		for _, datum := range downloadData.Data {
			if !datum.IsCanDL {
				continue
			}
			if err := downloader.Download(datum, _stream, path); err != nil {
				errors = append(errors, err)
			}
		}

		if len(errors) > 0 {
			fmt.Println("GetColumn errors:", errors[0])
			return
		}

		// 如果是专栏，则需要打印内容
		if course.IsColumn() {
			path, err := utils.Mkdir(utils.FileName(course.ColumnTitle, ""), "PDF")
			if err != nil {
				fmt.Println("GetColumn Mkdir:", err)
				return
			}
			cookies := application.LoginedCookies()
			for _, datum := range downloadData.Data {
				if !datum.IsCanDL {
					continue
				}
				err, exist := downloader.PrintToPDF(datum, cookies, path)
				if err != nil {
					errors = append(errors, err)
				}

				if !exist {
					time.Sleep(5 * time.Second)
				} else {
					have = true
				}
			}
		}

		if len(errors) > 0 {
			fmt.Println("GetColumn Mkdir:", errors[0])
			return
		}
		if !have {
			time.Sleep(30 * time.Second)
			have = false
		}

	}
}
