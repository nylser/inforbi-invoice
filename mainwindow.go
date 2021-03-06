package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/inforbi/inforbi-invoice/data"
	"github.com/skratchdot/open-golang/open"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type MainWindow struct {
	widgets.QMainWindow

	clientName   *widgets.QLabel
	clientChoose *widgets.QPushButton
	clientEdit   *widgets.QPushButton
	clientCreate *widgets.QPushButton

	invoiceName   *widgets.QLabel
	invoiceChoose *widgets.QPushButton
	invoiceEdit   *widgets.QPushButton
	invoiceCreate *widgets.QPushButton

	previewBtn   *widgets.QPushButton
	saveTexBtn   *widgets.QPushButton
	savePdfBtn   *widgets.QPushButton
	useRemoteBox *widgets.QCheckBox
}

var (
	clientSelected  = false
	invoiceSelected = false
	selectedClient  data.Client
	selectedInvoice data.Invoice
	lastOpenedPath  = ""
)

func initMainWindow() *MainWindow {
	this := NewMainWindow(nil, 0)
	this.SetWindowTitle(core.QCoreApplication_ApplicationName())
	upperWidget := widgets.NewQWidget(this, 0)
	lowerWidget := widgets.NewQWidget(nil, 0)
	upperGrid := widgets.NewQGridLayout(upperWidget)
	lowerGrid := widgets.NewQGridLayout(lowerWidget)
	upperGrid.SetSpacing(10)
	lowerGrid.SetSpacing(50)

	this.SetCentralWidget(upperWidget)

	upperGrid.AddWidget3(lowerWidget, 2, 0, 1, 4, core.Qt__AlignCenter)

	this.clientName = widgets.NewQLabel2("<No client selected>", nil, 0)
	this.clientChoose = widgets.NewQPushButton2("Choose...", nil)
	this.clientEdit = widgets.NewQPushButton2("Edit", nil)
	this.clientCreate = widgets.NewQPushButton2("Create", nil)

	this.clientChoose.ConnectPressed(this.chooseClient)
	this.clientEdit.ConnectPressed(this.editClient)
	this.clientCreate.ConnectPressed(this.createClient)

	upperGrid.AddWidget2(this.clientName, 0, 0, core.Qt__AlignLeft)
	upperGrid.AddWidget2(this.clientChoose, 0, 1, core.Qt__AlignLeft)
	upperGrid.AddWidget2(this.clientEdit, 0, 2, core.Qt__AlignLeft)
	upperGrid.AddWidget2(this.clientCreate, 0, 3, core.Qt__AlignLeft)

	this.invoiceName = widgets.NewQLabel2("<No invoice selected>", nil, 0)
	this.invoiceChoose = widgets.NewQPushButton2("Choose...", nil)
	this.invoiceEdit = widgets.NewQPushButton2("Edit", nil)
	this.invoiceCreate = widgets.NewQPushButton2("Create", nil)

	upperGrid.AddWidget2(this.invoiceName, 1, 0, core.Qt__AlignLeft)
	upperGrid.AddWidget2(this.invoiceChoose, 1, 1, core.Qt__AlignLeft)
	upperGrid.AddWidget2(this.invoiceEdit, 1, 2, core.Qt__AlignLeft)
	upperGrid.AddWidget2(this.invoiceCreate, 1, 3, core.Qt__AlignLeft)

	this.invoiceChoose.ConnectPressed(this.chooseInvoice)
	this.invoiceEdit.ConnectPressed(this.editInvoice)
	this.invoiceCreate.ConnectPressed(this.createInvoice)

	this.previewBtn = widgets.NewQPushButton2("Preview", nil)
	this.saveTexBtn = widgets.NewQPushButton2("Save .tex", nil)
	this.savePdfBtn = widgets.NewQPushButton2("Save .pdf", nil)
	this.useRemoteBox = widgets.NewQCheckBox2("Use remote-server for rendering", nil)
	this.useRemoteBox.SetChecked(true)

	this.previewBtn.ConnectPressed(this.preview)
	this.savePdfBtn.ConnectPressed(this.savePDF)
	this.saveTexBtn.ConnectPressed(this.saveTex)

	lowerGrid.SetSpacing(2)
	lowerGrid.AddWidget3(this.useRemoteBox, 0, 0, 1, 3, core.Qt__AlignCenter)
	lowerGrid.AddWidget2(this.previewBtn, 1, 0, core.Qt__AlignCenter)
	lowerGrid.AddWidget2(this.saveTexBtn, 1, 1, core.Qt__AlignCenter)
	lowerGrid.AddWidget2(this.savePdfBtn, 1, 2, core.Qt__AlignCenter)

	this.updateInvoice()
	this.updateClient()
	this.updateBottomBtns()

	return this
}

func (window *MainWindow) createClient() {
	selectedClient = data.Client{}
	window.editClient()
}

func (window *MainWindow) createInvoice() {
	if clientSelected {
		if invoiceSelected {
			var reply widgets.QMessageBox__StandardButton
			reply = widgets.QMessageBox_Question(window, "Create new invoice based on old one?", "Create new invoice based on old one?",
				widgets.QMessageBox__Yes|widgets.QMessageBox__No, widgets.QMessageBox__Yes)
			if reply == widgets.QMessageBox__Yes {
				selectedInvoice.Date = time.Now().Format("2006-01-02")
				file := selectedInvoice.GetFile()
				newfile := strings.Replace(file, strconv.Itoa(selectedInvoice.Number),
					strconv.Itoa(selectedInvoice.Number+1), 1)
				if file == newfile {
					ext := filepath.Ext(file)
					file = strings.Replace(file, ext, "_new"+ext, 1)
				} else {
					file = newfile
				}
				selectedInvoice.SetFile(file)
				selectedInvoice.Number = selectedInvoice.Number + 1
				selectedInvoice = selectedInvoice
			} else {
				selectedInvoice = data.Invoice{Date: time.Now().Format("2006-01-02")}
			}
		}

		window.editInvoice()
	}
}

func (window *MainWindow) chooseClient() {
	openPath := lastOpenedPath
	if len(lastOpenedPath) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			widgets.NewQErrorMessage(window).ShowMessage("Can't get directory!")
		}
		openPath = wd
	}
	dialog := widgets.NewQFileDialog(window, 0)
	path := dialog.GetOpenFileName(window, "Choose client", openPath,
		"*.json", "", 0)
	if len(path) > 0 {
		lastOpenedPath = filepath.Dir(path)
		client, err := data.DecodeClient(path)
		if err != nil {
			widgets.NewQErrorMessage(window).ShowMessage("Your selected file doesn't seem to be valid!")
		} else {
			selectedClient = client
			clientSelected = true
			window.updateClient()
			window.updateInvoice()
		}
	}
}

func (window *MainWindow) editClient() {
	cw := initClientEditDialog(selectedClient, window.ParentWidget())
	cw.Exec()
	if cw.Result() == 0 {
		selectedClient = cw.client
		window.updateClient()
		if !clientSelected {
			clientSelected = true
		}
	} else {
		selectedClient.EncodeClient()
	}

}

func (window *MainWindow) editInvoice() {
	iw := initInvoiceEditDialog(selectedInvoice, window.ParentWidget())
	iw.Exec()
	if iw.Result() == 0 {
		selectedInvoice = iw.invoice
		window.updateInvoice()
		if !invoiceSelected {
			invoiceSelected = true
		}
	} else {
		selectedInvoice.EncodeInvoice()
	}
}

func (window *MainWindow) chooseInvoice() {
	openPath := lastOpenedPath
	if len(lastOpenedPath) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			widgets.NewQErrorMessage(window).ShowMessage("Can't get directory!")
		}
		openPath = wd
	}
	dialog := widgets.NewQFileDialog(window, 0)
	path := dialog.GetOpenFileName(window, "Choose invoice", openPath,
		"*.json", "", 0)
	if len(path) > 0 {
		lastOpenedPath = filepath.Dir(path)
		invoice, err := data.DecodeInvoice(path)
		if err != nil {
			widgets.NewQErrorMessage(window).ShowMessage("Your selected file doesn't seem to be valid!")
		} else {
			selectedInvoice = invoice
			invoiceSelected = true
			window.updateInvoice()
		}
	}
}

func (window *MainWindow) updateInvoice() {
	if clientSelected && invoiceSelected {
		window.invoiceName.SetText("<" + strconv.Itoa(selectedInvoice.Number) + "> " + selectedInvoice.Project)
	}
	window.invoiceEdit.SetEnabled(clientSelected && invoiceSelected)
	window.invoiceChoose.SetEnabled(clientSelected)
	window.invoiceCreate.SetEnabled(clientSelected)
	window.updateBottomBtns()
}

func (window *MainWindow) updateClient() {
	if clientSelected {
		window.clientName.SetText(selectedClient.Name)
	}
	window.clientEdit.SetEnabled(clientSelected)
	window.updateBottomBtns()
}

func (window *MainWindow) updateBottomBtns() {
	condition := invoiceSelected && clientSelected
	window.previewBtn.SetEnabled(condition)
	window.savePdfBtn.SetEnabled(condition)
	window.saveTexBtn.SetEnabled(condition)
}

func (window *MainWindow) generateLatex() string {
	if clientSelected && invoiceSelected {
		file := core.NewQFile2(":/common/invoice.pylat")
		file.Open(core.QIODevice__ReadOnly)
		template := file.ReadAll().Data()
		template = selectedClient.ReplaceTemplate(template)
		template = selectedInvoice.ReplaceTemplate(template)
		return template
	} else {
		return ""
	}
}

func (window *MainWindow) preview() {
	dir, err := ioutil.TempDir("", "preview")
	tmpPDF := filepath.Join(dir, "preview.pdf")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if window.useRemoteBox.IsChecked() {
			err := window.remoteRender(tmpPDF)
			if err != nil {
				println(err)
			}
			open.Run(tmpPDF)
			go func() {
				time.Sleep(1 * time.Second)
				os.RemoveAll(dir)
			}()

		} else {
			window.localRender(dir, tmpPDF)
			open.Run(tmpPDF)
			go func() {
				time.Sleep(1 * time.Second)
				os.RemoveAll(dir)
			}()
		}
	}()
}

func (window *MainWindow) savePDF() {
	wd, err := os.Getwd()

	if err != nil {
		widgets.NewQErrorMessage(window).ShowMessage("Can't get directory!")
	}

	if len(lastOpenedPath) == 0 {
		lastOpenedPath = wd
	}

	dialog := widgets.NewQFileDialog(window, 0)
	path := dialog.GetSaveFileName(window, "Save invoice", lastOpenedPath,
		"*.pdf", "*.pdf", 0)

	if len(path) == 0 {
		widgets.NewQErrorMessage(window).ShowMessage("Can't save file without selected destination!")
		return
	}

	go func() {
		if window.useRemoteBox.IsChecked() {
			window.remoteRender(path)
		} else {
			dir, err := ioutil.TempDir("", "save")
			if err != nil {
				log.Fatal(err)
			}
			window.localRender(dir, path)
			os.RemoveAll(dir)
		}
	}()
}

func (window *MainWindow) saveTex() {
	wd, err := os.Getwd()
	if err != nil {
		widgets.NewQErrorMessage(window).ShowMessage("Can't get directory!")
	}
	if len(lastOpenedPath) == 0 {
		lastOpenedPath = wd
	}
	dialog := widgets.NewQFileDialog(window, 0)
	path := dialog.GetSaveFileName(window, "Save latex", lastOpenedPath,
		"*.tex", "*.tex", 0)
	if len(path) == 0 {
		widgets.NewQErrorMessage(window).ShowMessage("Can't save file without selected destination!")
		return
	}
	latex := window.generateLatex()
	err = ioutil.WriteFile(path, []byte(latex), 0644)
	if err != nil {
		log.Fatal(err)
		widgets.NewQErrorMessage(window).ShowMessage("Couldn't save file! " + err.Error())
	}

}

func (window *MainWindow) localRender(target_dir string, target_file string) {
	latex := window.generateLatex()
	tmplat := filepath.Join(target_dir, "render.tex")
	os.Mkdir(filepath.Join(target_dir, "Fonts"), 0777)
	list := []string{"/common/Fonts/FontAwesome.otf", "/common/Fonts/OpenSans-Bold.ttf",
		"/common/Fonts/OpenSans-Italic.ttf", "/common/Fonts/OpenSans-LightItalic.ttf", "/common/Fonts/OpenSans-Regular.ttf",
		"/common/dapper-invoice.cls"}

	for _, file := range list {
		target := filepath.Join(target_dir, strings.TrimPrefix(file, "/common/"))
		qFile := core.NewQFile2(":" + file)
		qFile.Open(core.QIODevice__ReadOnly)
		qFile.Copy(target)
		print(target)
	}

	print(target_dir)

	err := ioutil.WriteFile(tmplat, []byte(latex), 0644)
	if err != nil {
		log.Fatal(err)
		widgets.NewQErrorMessage(window).ShowMessage("Couldn't save file! " + err.Error())
	} else {
		command := exec.Command("xelatex", "-synctex=1", "-interaction=nonstopmode", "render.tex")
		command.Dir = target_dir
		out, err := command.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
		outstr := string(out)
		if strings.Contains(strings.ToLower(outstr), "rerun") {
			command := exec.Command("xelatex", "-synctex=1", "-interaction=nonstopmode", "render.tex")
			command.Dir = target_dir
			command.Run()
		}
		data.CopyFile(filepath.Join(target_dir, "render.pdf"), target_file)

	}
}

func (window *MainWindow) remoteRender(target_file string) error {
	latex := window.generateLatex()
	conn, err := net.Dial("tcp", "mineguild.net:7714")
	//conn, err := net.Dial("tcp", "localhost:7714")
	if err != nil {
		println(err)
		return err
	}
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	writer.WriteString("begin_send" + strconv.Itoa(len(latex)) + "\n")
	writer.Write([]byte(latex))
	writer.Flush()
	response, err := reader.ReadString('\n')
	if err != nil {
		widgets.NewQErrorMessage(window).ShowMessage("Error communicating!")
		return err
	}
	if strings.Trim(response, "\n") == "success" {
		writer.WriteString(selectedInvoice.Date + "\n")
		writer.Flush()
		response, err = reader.ReadString('\n')
		if strings.HasPrefix(response, "begin_send") {
			response = strings.TrimPrefix(response, "begin_send")
			response = strings.Trim(response, "\n")
			ulen, err := strconv.ParseInt(response, 10, 64)
			if err != nil {
				return err
			}
			ilen := int(ulen)
			file := data.ReceiveBlob(reader, ilen)
			if len(file) == ilen {
				writer.WriteString("success\n")
				err = ioutil.WriteFile(target_file, file, 0644)
			}
		}
	}
	return nil
}
