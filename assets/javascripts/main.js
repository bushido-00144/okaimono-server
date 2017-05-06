class User {
    constructor(userId) {
        this.userId = userId;
    }

    getUserInfo() {
        return fetch('/user/'+this.userId).then((res)=>{
            return res.json();
        }).then((userData)=>{
            this.userName = userData.UserName;
            return this.getUserRemainder();
        }).then((remainderData)=>{
            this.remainder = remainderData.Remainder;
        });
    }

    getUserRemainder() {
        return fetch('/remainder/'+this.userId).then((data)=>{
            return data.json();
        }).then((remainderData)=>{
            return remainderData;
        });
    }
}

let postJSON = (url, data)=>{
    return fetch(url, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    }).then((res)=>{
        return res.json();
    });
}

// DOM Loaded
document.addEventListener('DOMContentLoaded', ()=>{
    let user = new User(1);
    let promise = new Promise((resolve, reject)=>{
        resolve(user.getUserInfo());
    }).then(()=>{
        document.getElementById('user-name').innerHTML = user.userName;
        document.getElementById('user-remainder').innerHTML = user.remainder;

        document.getElementById('user-setting-button').addEventListener('click', ()=>{
            // 残高チャージ
            vex.dialog.prompt({
                message: '残高をチャージします',
                placeholder: '金額を入力してください',
                callback: (value)=>{
                    let amount = Number(value);
                    if(isNaN(amount)) {
                        console.log('Invalid value');
                        return;
                    }
                    postJSON('/remainder/charge', {Remainder: amount, UserID: user.userId})
                    .then((data)=>{
                        console.log(data);
                    });
                }
            });
        });
    });
})
