package ui

import (
  "fmt"
  "github.com/askovpen/goated/lib/msgapi"
  "github.com/jroimartin/gocui"
  "strconv"
  "log"
)

func getAreaNew(m msgapi.AreaPrimitive) string {
  if m.GetCount()-m.GetLast()>0 {
    return "\033[37;1m+\033[0m"
  } else {
    return " "
  }
}
func areaNext(g *gocui.Gui, v *gocui.View) error {
  if v != nil {
    cx, cy := v.Cursor()
    if cy==len(msgapi.Areas) {
      return nil
    }
    if err := v.SetCursor(cx, cy+1); err != nil {
      ox, oy := v.Origin()
      if cy+oy==len(msgapi.Areas) {
        return nil
      }
      if err := v.SetOrigin(ox, oy+1); err != nil {
        return err
      }
    }
  }
  return nil
}

func areaPrev(g *gocui.Gui, v *gocui.View) error {
  if v != nil {
    ox, oy := v.Origin()
    cx, cy := v.Cursor()
    log.Printf("cy: %d, oy: %d",cy,oy)
    if cy>1 {
      if err := v.SetCursor(cx, cy-1); err != nil  {
        log.Print(err)
        return err
      }
    } else if oy>0 {
      if err := v.SetOrigin(ox, oy-1); err != nil {
        log.Print(err)
        return err
      }
    }
  }
  return nil
}

func viewArea(g *gocui.Gui, v *gocui.View) error {
    _, oy := v.Origin()
    _, cy := v.Cursor()
    log.Printf("view %d",oy+cy)
    viewMsg(cy+oy-1,msgapi.Areas[cy+oy-1].GetLast())
    if _, err := g.SetCurrentView("MsgHeader"); err != nil {
      log.Print(err)
      return err
    }
    App.SetCurrentView("MsgHeader")
    App.SetCurrentView("MsgBody")
    return nil
}

func CreateAreaList() error {
  maxX, maxY := App.Size()
  AreaList, err:= App.SetView("AreaList", 0, 0, maxX-1, maxY-2);
  if err!=nil && err!=gocui.ErrUnknownView { 
    return err
  }
  AreaList.Wrap=false
  AreaList.Highlight = true
  AreaList.SelBgColor = gocui.ColorBlue
  AreaList.SelFgColor = gocui.ColorWhite | gocui.AttrBold
  AreaList.Clear()
  fmt.Fprintf(AreaList, "\033[33;1m Area %-"+strconv.FormatInt(int64(maxX-23),10)+"s %6s %6s \033[0m\n",
    "EchoID","Msgs","New")
  for i, a := range msgapi.Areas {
    fmt.Fprintf(AreaList, "%4d%s %-"+strconv.FormatInt(int64(maxX-23),10)+"s %6d %6d \n",
      i+1,
      getAreaNew(a),
      a.GetName(),
      a.GetCount(),
      a.GetCount()-a.GetLast())
  }
//  AreaList.SetCursor(0,1)
/*
  if _, err = setCurrentViewOnTop("AreaList"); err != nil {
    return err
  }
*/
   App.SetCurrentView("AreaList")
  _, cy := AreaList.Cursor()
  if cy==0 {
   areaNext(App,AreaList)
  }
//  if err := App.SetKeybinding("AreaList", gocui.KeyArrowDown, gocui.ModNone, AreaNext); err != nil {
//    return err
//  }
  return nil
}
