import React, { useState, useEffect } from "react";
import { MdOutlineAddBox } from "react-icons/md";
import { Button } from "@/Components/ui/button";
import { ConfirmModal } from "@/Components/Chatroom/ConfirmModal";
import { Spinner } from "@/Components/Spinner/Spinner";

export const Sidebar = ({
  chatrooms,
  loadingChatrooms,
  currChatroom,
  setCurrChatroom,
  setAddNewChat,
  showContent,
  onDeleteChatroom,
  unreadCounts,
  setUnreadCounts,
}) => {
  const [chatToDelete, setChatToDelete] = useState(null);

  const handleDeleteClick = (room) => {
    setChatToDelete(room);
  };

  const confirmDeletion = () => {
    if (typeof onDeleteChatroom === "function" && chatToDelete) {
      onDeleteChatroom(chatToDelete._id);
    }
    setChatToDelete(null);
  };

  const cancelDeletion = () => {
    setChatToDelete(null);
  };

  return (
    <div className="w-80 bg-green-600 text-white p-5 box-border overflow-y-auto h-full shadow-[2px_0_5px_rgba(0,0,0,0.1)]">
      {showContent && (
        <>
          <div className="flex justify-between">
            <h3 className="mt-0 text-3xl font-bold">Chats</h3>
            <MdOutlineAddBox
              size={35}
              className="cursor-pointer"
              onClick={() => setAddNewChat(true)}
            />
          </div>
          <div className="mt-2">
            {loadingChatrooms ? (
              <Spinner />
            ) : !chatrooms || chatrooms.length === 0 ? (
              "No chatrooms to show"
            ) : (
              chatrooms.filter(Boolean).map((room) => (
                <div key={room._id} className="flex items-center">
                  <Button
                    variant={currChatroom === room ? "selected" : "inverse"}
                    className="flex-1 my-2 text-base relative"
                    onClick={() => {
                      setCurrChatroom(room);
                      setUnreadCounts((prev) => ({ ...prev, [room._id]: 0 }));
                    }}
                  >
                    {room.name ?? "Unnamed Chat"}{" "}
                    {Number(unreadCounts[room._id]) > 0 && (
                        <span className="bg-red-600 text-neutral-100 p-px text-sm absolute -top-2 -right-1 flex items-center justify-center rounded-full min-w-[1.25rem] h-5 px-1">{Number(unreadCounts[room._id])}</span>
                    )}
                  </Button>
                  <Button
                    variant="destructive"
                    className="ml-2"
                    onClick={() =>
                      handleDeleteClick({
                        _id: room._id,
                        name: room.name ?? "Unnamed Chat",
                      })
                    }
                  >
                    Delete
                  </Button>
                </div>
              ))
            )}
          </div>
        </>
      )}
      {chatToDelete && (
        <ConfirmModal
          message={`Are you sure you want to delete the chatroom "${chatToDelete.name}"? This will delete all messages for everyone.`}
          onConfirm={confirmDeletion}
          onCancel={cancelDeletion}
        />
      )}
    </div>
  );
};
