/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-03-22 13:53:38
* @brief:
*
**/
package main

import "db"
import "record"
import "base/xnet"
import . "logger"

func ProtoString(str string) *string {
	return &str
}

type CharacterService struct {
}

func (*CharacterService) Update(ctx *xnet.Context, in *record.UpdateCharacterRequest, out *record.UpdateCharacterResponse) *string {

	in.Data.Recover(in.GetMask())

	DEBUG("update by: %v", in.Data.Key())

	if in.Data.Flush() {
		return nil
	}

	return ProtoString("update failed.")
}

type RecordCharacterWrapper struct {
	*record.RecordCharacter
}

func (this *RecordCharacterWrapper) FlushTo(service record.ICharacterServiceClient) {
	// no changes.
	if 0 == this.Mask() {
		return
	}

	// build update request.
	//
	request := &record.UpdateCharacterRequest{
		Data: &record.RecordCharacter{},
		Mask: this.Mask(),
	}

	// merge changes to request.
	request.Data.CopyFrom(this.Mask(), this.RecordCharacter)

	// reset dirty mask.
	this.ClearMask()

	service.AsyncUpdate(request, func(err *string, resp *record.UpdateCharacterResponse) {
		if err != nil {
			ERROR("update RecordCharacter failed. %v %s", this.Key(), *err)
			// recover mask.
			this.Recover(request.Mask)
		}
	})
}

func main() {
	xnet.Assert(db.InitializeMongodb("mongodb://192.168.1.2:27017/Test1"), "fail to connect mgo.")

	// db service.
	{
		service, err := record.NewCharacterServiceImpl("tcp://127.0.0.1:3333", &CharacterService{})
		xnet.Assert(nil == err, err)

		go service.AcceptLoop(func(session xnet.ISession) bool {
			return true
		})
	}

	client, err := record.DialCharacterService("tcp://127.0.0.1:3333")
	xnet.Assert(nil == err, err)

	wrapper := &RecordCharacterWrapper{&record.RecordCharacter{Uid: 1111}}

	record.Truncate(wrapper.Table())
	record.Insert(wrapper.RecordCharacter)

	wrapper.Level = 20
	wrapper.Role = 1003
	wrapper.Dirty(record.RecordCharacterMask_Level, record.RecordCharacterMask_Role)

	wrapper.FlushTo(client)

	select {}
}
